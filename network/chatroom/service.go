package chatroom

import (
	"context"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync"
	"time"
	"whitenoise/common/account"
	"whitenoise/common/config"
	"whitenoise/common/log"
	"whitenoise/internal/pb"
	"whitenoise/network/session"
	"whitenoise/protocol/ack"
)

const NewRoomProtocol = "/roomnew"
const GetRoomProtocol = "/roomget"
const MaxRoomsForOneClient = 1
const MaxKeywordsSize = 5
const MaxKeywordsLength = 10
const MaxNameLength = 10

type RoomInfo struct {
	Name      string
	Keywords  []string
	StartTime time.Time
	Hoster    string
	StreamID  string
	PeerID    peer.ID
}

type RoomDnsService struct {
	host        host.Host
	rooms       sync.Map //key: roomname string v:roominfo
	actorCtx    *actor.RootContext
	pid         *actor.PID
	ctx         context.Context
	ackPid      *actor.PID
	clientRooms sync.Map
}

func NewDnsService(parent context.Context, actCtx *actor.RootContext, cfg *config.NetworkConfig, h core.Host) *RoomDnsService {
	ctx, _ := context.WithCancel(parent)
	service := RoomDnsService{
		host:        h,
		rooms:       sync.Map{},
		actorCtx:    actCtx,
		ctx:         ctx,
		ackPid:      nil,
		clientRooms: sync.Map{},
	}
	return &service
}

func (dns *RoomDnsService) Start() {
	props := actor.PropsFromProducer(func() actor.Actor {
		return dns
	})
	dns.pid = dns.actorCtx.Spawn(props)
	dns.host.SetStreamHandler(NewRoomProtocol, dns.NewRoomStreamHandler)
	dns.host.SetStreamHandler(GetRoomProtocol, dns.GetRoomStreamHandler)
}

func (dns *RoomDnsService) SetPid(ackPid *actor.PID) {
	dns.ackPid = ackPid
}

func (dns *RoomDnsService) Pid() *actor.PID {
	return dns.pid
}

func (dns *RoomDnsService) NewRoomStreamHandler(stream network.Stream) {
	str := session.NewStream(stream, dns.ctx)
	payloadBytes, err := str.RW.ReadMsg()
	if err != nil {
		return
	}

	var newRoom = pb.NewRoom{}
	err = proto.Unmarshal(payloadBytes, &newRoom)
	if err != nil {
		log.Error("unmarshal err", err)
	}

	if newRoom.From == "" || newRoom.Name == "" || newRoom.Commandid == "" || len(newRoom.Keywords) > MaxKeywordsSize || len(newRoom.Name) > MaxNameLength {
		log.Debug("new room request miss field")
		return
	}

	for _, keyword := range newRoom.Keywords {
		if len(keyword) > MaxKeywordsLength {
			return
		}
	}

	_, err = account.WhiteNoiseIDfromString(newRoom.From)
	if err != nil {
		return
	}

	ackMsg := pb.Ack{
		CommandId: newRoom.Commandid,
		Result:    false,
		Data:      []byte{},
	}
	reqAck := ack.ReqAck{
		Ack:    &ackMsg,
		PeerId: str.RemotePeer,
	}

	_, ok := dns.clientRooms.Load(str.RemotePeer)
	if ok {
		ackMsg.Data = []byte(fmt.Sprintf("Can't start more than %v chatrooms", MaxRoomsForOneClient))
		dns.actorCtx.Request(dns.ackPid, reqAck)
		return
	}

	_, ok = dns.rooms.Load(newRoom.Name)
	if ok {
		ackMsg.Data = []byte(fmt.Sprintf("Name %v used", newRoom.Name))
		dns.actorCtx.Request(dns.ackPid, reqAck)
		return
	}

	dns.rooms.Store(newRoom.Name, RoomInfo{
		Name:      newRoom.Name,
		Keywords:  newRoom.Keywords,
		StartTime: time.Now(),
		Hoster:    newRoom.From,
		StreamID:  str.StreamId,
		PeerID:    str.RemotePeer,
	})

	dns.clientRooms.Store(str.RemotePeer, newRoom.Name)

	//ack success
	ackMsg.Result = true
	dns.actorCtx.Request(dns.ackPid, reqAck)
}

func (dns *RoomDnsService) Offline(peerID peer.ID) {
	if v, ok := dns.clientRooms.Load(peerID); ok {
		name := v.(string)
		dns.DeleteRoom(name)
	}
}

func (dns *RoomDnsService) DeleteRoom(name string) {
	v, ok := dns.rooms.Load(name)
	if !ok {
		return
	}
	info := v.(RoomInfo)
	dns.rooms.Delete(name)
	//todo:Now One WhiteNoiseID One Room, extend later.
	dns.clientRooms.Delete(info.PeerID)
}

func (dns *RoomDnsService) GetRoomStreamHandler(stream core.Stream) {
	defer stream.Close()
	str := session.NewStream(stream, dns.ctx)
	payloadBytes, err := str.RW.ReadMsg()
	if err != nil {
		return
	}

	var newRoom = pb.GetRoom{}
	err = proto.Unmarshal(payloadBytes, &newRoom)
	if err != nil {
		log.Error("unmarshal err", err)
	}

	ackMsg := pb.Ack{
		CommandId: newRoom.CommandId,
		Result:    true,
		Data:      []byte{},
	}

	roomList := make([]*pb.RoomInfo, 0)

	dns.rooms.Range(func(key interface{}, v interface{}) bool {
		name := key.(string)
		info := v.(RoomInfo)
		roomList = append(roomList, &pb.RoomInfo{
			Name:      name,
			Keywords:  info.Keywords,
			Starttime: info.StartTime.String(),
			Hoster:    info.Hoster,
		})
		return true
	})

	resRooms := pb.ResRooms{RoomList: roomList}
	data, err := proto.Marshal(&resRooms)
	if err != nil {
		log.Error(err)
		return
	}

	ackMsg.Data = data

	dns.actorCtx.Request(dns.ackPid, ack.ReqAck{
		Ack:    &ackMsg,
		PeerId: stream.Conn().RemotePeer(),
	})
}
