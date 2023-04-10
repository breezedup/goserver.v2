package main

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/breezedup/goserver.v2/core/netlib"
	"github.com/breezedup/goserver.v2/examples/protocol"
)

type CSPacketPingPacketFactory struct {
}

type CSPacketPingHandler struct {
}

func (this *CSPacketPingPacketFactory) CreatePacket() interface{} {
	pack := &protocol.CSPacketPing{}
	return pack
}

func (this *CSPacketPingHandler) Process(session *netlib.Session, packetid int, data interface{}) error {
	if ping, ok := data.(*protocol.CSPacketPing); ok {
		pong := &protocol.SCPacketPong{
			TimeStamb: proto.Int64(ping.GetTimeStamb()),
			Message:   ping.GetMessage(),
		}
		proto.SetDefaults(pong)
		session.Send(int(protocol.PacketID_PACKET_SC_PONG), pong)
	}
	return nil
}

func init() {
	netlib.RegisterHandler(int(protocol.PacketID_PACKET_CS_PING), &CSPacketPingHandler{})
	netlib.RegisterFactory(int(protocol.PacketID_PACKET_CS_PING), &CSPacketPingPacketFactory{})
}
