package signal

import (
	"github.com/breezedup/goserver.v2/core"
)

var Config = Configuration{}

type Configuration struct {
	SupportSignal bool
}

func (c *Configuration) Name() string {
	return "signal"
}

func (c *Configuration) Init() error {
	if c.SupportSignal {
		//demon goroutine
		go SignalHandlerModule.ProcessSignal()
	}
	return nil
}

func (c *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
