package netlib

import (
	"github.com/breezedup/goserver.v2/core"
	"github.com/breezedup/goserver.v2/core/basic"
)

type startIoService struct {
	sc *SessionConfig
}

func (sis *startIoService) Done(o *basic.Object) error {

	s := NetModule.newIoService(sis.sc)
	if s != nil {
		NetModule.pool[sis.sc.Id] = s
		s.start()
	}

	return nil
}

func SendStartNetIoService(sc *SessionConfig) bool {
	return core.CoreObject().SendCommand(&startIoService{sc: sc}, false)
}
