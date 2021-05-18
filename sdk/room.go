package sdk

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"time"
	"whitenoise/common/log"
	"whitenoise/internal/pb"
	"whitenoise/secure"
)

type BasicRoom struct {
	client        Client
	name          string
	keywords      []string
	broadcastChan chan []byte //broadcast to peers
	msgChan       chan []byte //display in ui
}

func NewRoom(parent context.Context, name string, keywords []string) (*BasicRoom, error) {
	ctx, _ := context.WithCancel(parent)
	client, err := NewClient(ctx)
	if err != nil {
		panic(err)
	}
	room := BasicRoom{
		client:        client,
		name:          name,
		keywords:      keywords,
		broadcastChan: make(chan []byte),
		msgChan:       make(chan []byte),
	}
	return &room, nil
}

func (r *BasicRoom) Start() {
	log.Debug("Regsiter proxy")
	err := r.RegisterProxy()
	if err != nil {
		panic(err)
	}

	log.Debug("Regsiter Room")
	err = r.RegisterRoom()
	if err != nil {
		panic(err)
	}

	r.SubNewCircuit()
	go r.BroadCastWorker()

	go func() {
		for {
			payload := <-r.msgChan
			msg := pb.ChatMessage{}
			err := proto.Unmarshal(payload, &msg)
			if err != nil {
				continue
			}
			log.Info(msg.Nick, ": ", string(msg.Data))
		}
	}()

}

//todo:enable reset name and keywords for room
//func (r *BasicRoom) SetName(name string) {
//	r.name = name
//}
//
//func (r *BasicRoom) SetKeyWords(keywords []string) {
//	r.keywords = keywords
//}

func (room *BasicRoom) RegisterProxy() error {
	if room.client == nil {
		return errors.New("no client")
	}
	//register to proxy
	peers, err := room.client.GetMainNetPeers(10)
	if err != nil {
		return err
	}

	l := len(peers)
	if l == 0 {
		return errors.New("get mainnet peers 0 length")
	}

	index := rand.New(rand.NewSource(time.Now().UnixNano())).Int() % l
	err = room.client.Register(peers[index])
	if err != nil {
		return err
	}
	return nil
}

func (room *BasicRoom) RegisterRoom() error {
	if room.name == "" || room.client == nil {
		return errors.New("room not init")
	}
	return room.client.RegNewRoom(room.name, room.keywords)
}

func (room *BasicRoom) GetName() string {
	return room.name
}

func (room *BasicRoom) GetKeywords() []string {
	return room.keywords
}

func (room *BasicRoom) SubNewCircuit() {
	err := room.client.EventBus().Subscribe(GetCircuitTopic, func(sessionID string) {
		go func() {
			room.HandleNewCircuit(sessionID)
		}()
	})
	if err != nil {
		panic(err)
	}
}

func (room *BasicRoom) HandleNewCircuit(sessionID string) {
	conn, ok := room.client.GetCircuit(sessionID)
	if !ok {
		return
	}
	for {
		payload, err := secure.ReadPayload(conn)
		if err != nil {
			//if err != io.EOF{
			//	return
			//}
			return
		}
		room.broadcastChan <- payload
		room.msgChan <- payload
	}
}

func (room *BasicRoom) BroadCastWorker() {
	go func() {
		for {
			payload := <-room.broadcastChan
			room.BroadCast(payload)
		}
	}()
}

func (room *BasicRoom) BroadCast(payload []byte) {
	sessionIdList := room.client.GetAllCircuits()
	for _, sessionID := range sessionIdList {
		conn, ok := room.client.GetCircuit(sessionID)
		if !ok {
			continue
		}
		_, err := conn.Write(secure.EncodePayload(payload))
		if err != nil {
			continue
		}
	}
}
