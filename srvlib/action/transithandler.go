package action

import (
	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/breezedup/goserver.v2/core/netlib"
	"github.com/breezedup/goserver.v2/srvlib"
	"github.com/breezedup/goserver.v2/srvlib/protocol"
)

type PacketTransitPacketFactory struct {
}

type PacketTransitHandler struct {
}

func init() {
	netlib.RegisterHandler(int(protocol.SrvlibPacketID_PACKET_SS_TRANSIT), &PacketTransitHandler{})
	netlib.RegisterFactory(int(protocol.SrvlibPacketID_PACKET_SS_TRANSIT), &PacketTransitPacketFactory{})
}

func (this *PacketTransitPacketFactory) CreatePacket() interface{} {
	pack := &protocol.SSPacketTransit{}
	return pack
}

func (this *PacketTransitHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	logger.Logger.Trace("PacketTransitHandler.Process")
	if pr, ok := data.(*protocol.SSPacketTransit); ok {
		targetS := srvlib.ServerSessionMgrSington.GetSession(int(pr.GetSArea()), int(pr.GetSType()), int(pr.GetSId()))
		if targetS != nil {
			targetS.Send(int(pr.GetPacketId()), pr.GetData())
		}
	}
	return nil
}
