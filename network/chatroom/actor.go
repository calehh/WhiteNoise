package chatroom

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"whitenoise/internal/actorMsg"
)

func (dns *RoomDnsService) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case actorMsg.ReqStreamClosed:
		dns.Offline(msg.PeerID)
	default:
		//log.Debugf("Gossip actor cannot handle this request %v", msg)
	}
}
