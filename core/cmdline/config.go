package cmdline

import (
	"github.com/breezedup/goserver.v2/core"
)

var Config = Configuration{}

type Configuration struct {
	SupportCmdLine bool
}

func (c *Configuration) Name() string {
	return "cmdline"
}

func (c *Configuration) Init() error {
	return nil
}

func (c *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
