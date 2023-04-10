package main

import (
	"time"

	"github.com/breezedup/goserver.v2/core"
	_ "github.com/breezedup/goserver.v2/core/builtin/action"
	_ "github.com/breezedup/goserver.v2/core/builtin/filter"
	"github.com/breezedup/goserver.v2/core/module"
	"github.com/breezedup/goserver.v2/core/netlib"
)

var (
	Config         = Configuration{}
	PressureModule = &PressureTest{}
	StartCnt       = 0
)

type Configuration struct {
	Count    int
	Connects netlib.SessionConfig
}

func (this *Configuration) Name() string {
	return "pressure"
}

func (this *Configuration) Init() error {
	this.Connects.Init()
	return nil
}

func (this *Configuration) Close() error {
	return nil
}

type PressureTest struct {
}

func (this PressureTest) ModuleName() string {
	return "pressure-module"
}

func (this *PressureTest) Init() {
	cfg := Config.Connects
	for i := 0; i < Config.Count; i++ {
		cfg.Id += i
		netlib.Connect(&cfg)
	}
}

func (this *PressureTest) Update() {
	return
}

func (this *PressureTest) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	core.RegistePackage(&Config)
	module.RegisteModule(PressureModule, time.Second*30, 50)
}
