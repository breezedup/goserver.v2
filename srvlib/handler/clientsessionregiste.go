// SessionHandlerClientRegiste
package handler

import (
	"github.com/breezedup/goserver.v2/core/netlib"
	"github.com/breezedup/goserver.v2/srvlib"
)

var (
	SessionHandlerClientRegisteName = "session-client-registe"
)

type SessionHandlerClientRegiste struct {
}

func (sfcr SessionHandlerClientRegiste) GetName() string {
	return SessionHandlerClientRegisteName
}

func (sfl *SessionHandlerClientRegiste) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Opened | 1<<netlib.InterestOps_Closed
}

func (sfl *SessionHandlerClientRegiste) OnSessionOpened(s *netlib.Session) {
	srvlib.ClientSessionMgrSington.RegisteSession(s)
}

func (sfl *SessionHandlerClientRegiste) OnSessionClosed(s *netlib.Session) {
	srvlib.ClientSessionMgrSington.UnregisteSession(s)
}

func (sfl *SessionHandlerClientRegiste) OnSessionIdle(s *netlib.Session) {
}

func (sfl *SessionHandlerClientRegiste) OnPacketReceived(s *netlib.Session, packetid int, packet interface{}) {
}

func (sfl *SessionHandlerClientRegiste) OnPacketSent(s *netlib.Session, data []byte) {
}

func init() {
	netlib.RegisteSessionHandlerCreator(SessionHandlerClientRegisteName, func() netlib.SessionHandler {
		return &SessionHandlerClientRegiste{}
	})
}
