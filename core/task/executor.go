package task

import (
	"fmt"

	"github.com/breezedup/goserver.v2/core"
	"github.com/breezedup/goserver.v2/core/basic"
	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/stathat/consistent"
)

var (
	WorkerIdGenerator int       = 0
	WorkerInitialCnt            = 8
	WorkerVirtualNum            = 8
	TaskExecutor      *Executor = NewExecutor()
)

type Executor struct {
	*basic.Object
	c          *consistent.Consistent
	workers    map[string]*Worker
	fixWorkers map[string]*Worker
}

func NewExecutor() *Executor {
	e := &Executor{
		c:          consistent.New(),
		workers:    make(map[string]*Worker),
		fixWorkers: make(map[string]*Worker),
	}

	return e
}

func (e *Executor) Start() {
	logger.Logger.Trace("Executor Start")
	defer logger.Logger.Trace("Executor Start [ok]")

	e.Object = basic.NewObject(core.ObjId_ExecutorId,
		"executor",
		Config.Options,
		nil)
	e.c.NumberOfReplicas = WorkerVirtualNum
	e.UserData = e
	e.addWorker(Config.Worker.WorkerCnt)

	core.LaunchChild(TaskExecutor.Object)
}

func (e *Executor) addWorker(workerCnt int) {
	for i := 0; i < workerCnt; i++ {
		w := &Worker{
			Object: basic.NewObject(WorkerIdGenerator,
				fmt.Sprintf("worker_%d", WorkerIdGenerator),
				Config.Worker.Options,
				nil),
		}
		WorkerIdGenerator++

		w.UserData = w
		e.LaunchChild(w.Object)
		e.c.Add(w.Name)
		e.workers[w.Name] = w
	}
}

func (e *Executor) getWorker(name string) *Worker {
	if w, exist := e.workers[name]; exist {
		return w
	}
	return nil
}

func (e *Executor) getFixWorker(name string) *Worker {
	if w, exist := e.fixWorkers[name]; exist {
		return w
	}
	return nil
}

func (e *Executor) addFixWorker(name string) *Worker {
	logger.Logger.Infof("Executor.AddFixWorker(%v)", name)
	w := &Worker{
		Object: basic.NewObject(WorkerIdGenerator,
			name,
			Config.Worker.Options,
			nil),
	}
	WorkerIdGenerator++

	w.UserData = w
	e.LaunchChild(w.Object)
	e.fixWorkers[name] = w
	return w
}
