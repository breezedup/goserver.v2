package module

import (
	"container/list"
	"time"

	"github.com/breezedup/goserver.v2/core"
	"github.com/breezedup/goserver.v2/core/basic"
	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/breezedup/goserver.v2/core/utils"
)

const (
	///module state
	ModuleStateInvalid int = iota
	ModuleStateInit
	ModuleStateRun
	ModuleStateShutdown
	ModuleStateWaitShutdown
	ModuleStateFini
	///other
	ModuleMaxCount = 1024
)

var (
	AppModule = newModuleMgr()
)

type ModuleEntity struct {
	lastTick     time.Time
	tickInterval time.Duration
	priority     int
	module       Module
	quited       bool
}

type PreloadModuleEntity struct {
	priority int
	module   PreloadModule
}

type ModuleMgr struct {
	*basic.Object
	modules       *list.List
	modulesByName map[string]*ModuleEntity
	preloadModule *list.List
	state         int
	waitShutAct   chan interface{}
	waitShutCnt   int
	waitShut      bool
	currTimeSec   int64
	currTimeNano  int64
}

func newModuleMgr() *ModuleMgr {
	mm := &ModuleMgr{
		modules:       list.New(),
		preloadModule: list.New(),
		modulesByName: make(map[string]*ModuleEntity),
		waitShutAct:   make(chan interface{}, ModuleMaxCount),
		state:         ModuleStateInvalid,
	}

	return mm
}

func (this *ModuleMgr) GetCurrTimeSec() int64 {
	return this.currTimeSec
}

func (this *ModuleMgr) GetCurrTimeNano() int64 {
	return this.currTimeNano
}

func (this *ModuleMgr) RegisteModule(m Module, tickInterval time.Duration, priority int) {
	logger.Logger.Infof("module [%16s] registe;interval=%v,priority=%v", m.ModuleName(), tickInterval, priority)
	mentiry := &ModuleEntity{
		lastTick:     time.Now(),
		tickInterval: tickInterval,
		priority:     priority,
		module:       m,
	}

	this.modulesByName[m.ModuleName()] = mentiry

	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok {
			if priority < me.priority {
				this.modules.InsertBefore(mentiry, e)
				return
			}
		}
	}
	this.modules.PushBack(mentiry)
}

func (this *ModuleMgr) UnregisteModule(m Module) {
	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok {
			if me.module == m {
				delete(this.modulesByName, m.ModuleName())
				this.modules.Remove(e)
				return
			}
		}
	}
}

func (this *ModuleMgr) RegistePreloadModule(m PreloadModule, priority int) {
	mentiry := &PreloadModuleEntity{
		priority: priority,
		module:   m,
	}

	for e := this.preloadModule.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*PreloadModuleEntity); ok {
			if priority < me.priority {
				this.preloadModule.InsertBefore(mentiry, e)
				return
			}
		}
	}
	this.preloadModule.PushBack(mentiry)
}

func (this *ModuleMgr) UnregistePreloadModule(m PreloadModule) {
	for e := this.preloadModule.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*PreloadModuleEntity); ok {
			if me.module == m {
				this.preloadModule.Remove(e)
				return
			}
		}
	}
}

func (this *ModuleMgr) Start() *utils.Waitor {
	logger.Logger.Info("Startup PreloadModules")
	for e := this.preloadModule.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*PreloadModuleEntity); ok {
			me.module.Start()
		}
	}
	logger.Logger.Info("Startup PreloadModules [ok]")

	this.Object = basic.NewObject(core.ObjId_CoreId,
		"core",
		Config.Options,
		this)
	this.UserData = this
	core.LaunchChild(this.Object)
	core.AppCtx.CoreObj = this.Object
	this.state = ModuleStateInit
	//给模块预留调度的空间，防止主线程直接跑过去
	select {
	case <-time.After(time.Second):
	}
	return this.Object.Waitor
}

func (this *ModuleMgr) Close() {
	if this.state != ModuleStateShutdown && this.state != ModuleStateWaitShutdown {
		this.state = ModuleStateShutdown
	}
}

func (this *ModuleMgr) init() {
	logger.Logger.Info("Start Initialize Modules")
	defer logger.Logger.Info("Start Initialize Modules [ok]")
	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok && !me.quited {
			logger.Logger.Infof("module [%16s] init...", me.module.ModuleName())
			me.module.Init()
			logger.Logger.Infof("module [%16s] init[ok]", me.module.ModuleName())
		}
	}
	this.state = ModuleStateRun
}

func (this *ModuleMgr) update() {
	nowTime := time.Now()
	this.currTimeSec = nowTime.Unix()
	this.currTimeNano = nowTime.UnixNano()
	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok && !me.quited {
			me.safeUpt(nowTime)
		}
	}
}

func (this *ModuleMgr) shutdown() {
	if this.waitShut {
		return
	}
	logger.Logger.Info("ModuleMgr shutdown()")
	this.waitShut = true
	this.state = ModuleStateWaitShutdown
	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok {
			logger.Logger.Infof("module [%16s] shutdown...", me.module.ModuleName())
			me.safeShutdown(this.waitShutAct)
			logger.Logger.Infof("module [%16s] shutdown[ok]", me.module.ModuleName())
			this.waitShutCnt++
		}
	}
}

func (this *ModuleMgr) checkShutdown() bool {
	select {
	case param := <-this.waitShutAct:
		logger.Logger.Infof("module [%16s] shutdowned", param)
		if name, ok := param.(string); ok {
			me := this.getModuleEntityByName(name)
			if me != nil && !me.quited {
				me.quited = true
				this.waitShutCnt--
			} else {
				logger.Logger.Infof("Module %v is already quit.", name)
			}
		} else {
			logger.Logger.Infof("Module shut down param %v error.", param)
		}
	default:
		logger.Infof("Left %v module wait shut down.", this.waitShutCnt)
	}
	if this.waitShutCnt == 0 {
		this.state = ModuleStateFini
		return true
	}
	this.update()
	return false
}

func (this *ModuleMgr) tick() {

	switch this.state {
	case ModuleStateInit:
		this.init()
	case ModuleStateRun:
		this.update()
	case ModuleStateShutdown:
		this.shutdown()
	case ModuleStateWaitShutdown:
		this.checkShutdown()
	case ModuleStateFini:
		this.fini()
	}
}

func (this *ModuleMgr) fini() {
	core.Terminate(this.Object)
	this.state = ModuleStateInvalid
	logger.Logger.Info("=============ModuleMgr fini=============")
	logger.Logger.Flush()
}

func (this *ModuleMgr) getModuleEntityByName(name string) *ModuleEntity {
	if me, exist := this.modulesByName[name]; exist {
		return me
	}
	return nil
}

func (this *ModuleMgr) GetModuleByName(name string) Module {
	if me, exist := this.modulesByName[name]; exist {
		return me.module
	}
	return nil
}

func (this *ModuleEntity) safeUpt(nowTime time.Time) {
	defer utils.DumpStackIfPanic("ModuleEntity.safeTick")

	if nowTime.Sub(this.lastTick) >= this.tickInterval {
		this.lastTick = nowTime
		this.module.Update()
	}
}

func (this *ModuleEntity) safeShutdown(shutWaitAck chan<- interface{}) {
	defer utils.DumpStackIfPanic("ModuleEntity.safeShutdown")
	this.module.Shutdown()
}

func (this *ModuleMgr) OnStart() {}
func (this *ModuleMgr) OnStop()  {}
func (this *ModuleMgr) OnTick() {
	this.tick()
}

func RegistePreloadModule(m PreloadModule, priority int) {
	AppModule.RegistePreloadModule(m, priority)
}

func RegisteModule(m Module, tickInterval time.Duration, priority int) {
	AppModule.RegisteModule(m, tickInterval, priority)
}

func UnregisteModule(m Module) {
	AppModule.waitShutAct <- m.ModuleName()
}

func Start() *utils.Waitor {
	err := core.ExecuteHook(core.HOOK_BEFORE_START)
	if err != nil {
		logger.Logger.Error("ExecuteHook(HOOK_BEFORE_START) error", err)
	}
	return AppModule.Start()
}

func Stop() {
	AppModule.Close()
	err := core.ExecuteHook(core.HOOK_AFTER_STOP)
	if err != nil {
		logger.Logger.Error("ExecuteHook(HOOK_BEFORE_START) error", err)
	}
}
