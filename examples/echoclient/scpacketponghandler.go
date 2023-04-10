package main

import (
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/breezedup/goserver.v2/core/netlib"
	"github.com/breezedup/goserver.v2/examples/protocol"
)

type SCPacketPongPacketFactory struct {
}

type SCPacketPongHandler struct {
}

func (this *SCPacketPongPacketFactory) CreatePacket() interface{} {
	pack := &protocol.SCPacketPong{}
	return pack
}

func (this *SCPacketPongHandler) Process(session *netlib.Session, packetid int, data interface{}) error {
	if pong, ok := data.(*protocol.SCPacketPong); ok {
		ping := &protocol.CSPacketPing{
			TimeStamb: proto.Int64(time.Now().Unix()),
			Message:   pong.GetMessage(),
		}
		proto.SetDefaults(ping)
		session.Send(int(protocol.PacketID_PACKET_CS_PING), ping)
	}
	return nil
}

func init() {
	netlib.RegisterHandler(int(protocol.PacketID_PACKET_SC_PONG), &SCPacketPongHandler{})
	netlib.RegisterFactory(int(protocol.PacketID_PACKET_SC_PONG), &SCPacketPongPacketFactory{})
}
