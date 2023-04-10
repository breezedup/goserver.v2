package task

import (
	"sync/atomic"
	"time"

	"github.com/breezedup/goserver.v2/core"
	"github.com/breezedup/goserver.v2/core/basic"
	"github.com/breezedup/goserver.v2/core/container"
	"github.com/breezedup/goserver.v2/core/container/recycler"
	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/breezedup/goserver.v2/core/utils"
)

type Callable interface {
	Call(*basic.Object) interface{}
}

type CompleteNotify interface {
	Done(interface{}, *Task)
}

type CallableWrapper func(o *basic.Object) interface{}

func (cw CallableWrapper) Call(o *basic.Object) interface{} {
	return cw(o)
}

type CompleteNotifyWrapper func(interface{}, *Task)

func (cnw CompleteNotifyWrapper) Done(i interface{}, t *Task) {
	cnw(i, t)
}

type Task struct {
	s          *basic.Object
	c          Callable
	n          CompleteNotify
	r          chan interface{}
	env        *container.SynchronizedMap
	tCreate    time.Time
	tStart     time.Time
	alertTime  time.Duration
	name       string
	refTaskCnt int32
}

func New(s *basic.Object, c Callable, n CompleteNotify, name ...string) *Task {
	t := &Task{
		s:       s,
		c:       c,
		n:       n,
		r:       make(chan interface{}, 1),
		tCreate: time.Now(),
	}
	if len(name) != 0 {
		t.name = name[0]
	}
	if s == nil {
		t.s = core.CoreObject()
	}
	return t
}

func (t *Task) AddRefCnt(cnt int32) int32 {
	return atomic.AddInt32(&t.refTaskCnt, cnt)
}

func (t *Task) GetRefCnt() int32 {
	return atomic.LoadInt32(&t.refTaskCnt)
}

func (t *Task) Get() interface{} {
	if t.n != nil {
		panic("Task result by CompleteNotify return")
	}

	return <-t.r
}

func (t *Task) GetWithTimeout(timeout time.Duration) interface{} {
	if timeout == 0 {
		return t.Get()
	} else {
		timer := recycler.GetTimer(timeout)
		defer recycler.GiveTimer(timer)
		select {
		case r, ok := <-t.r:
			if ok {
				return r
			} else {
				return nil
			}
		case <-timer.C:
			return nil
		}
	}
	return nil
}

func (t *Task) GetEnv(k interface{}) interface{} {
	if t.env == nil {
		return nil
	}
	return t.env.Get(k)
}

func (t *Task) PutEnv(k, v interface{}) bool {
	if t.env == nil {
		t.env = container.NewSynchronizedMap()
	}
	if t.env != nil {
		t.env.Set(k, v)
	}

	return true
}

func (t *Task) run(o *basic.Object) (e error) {
	defer utils.DumpStackIfPanic("Task::run")

	t.tStart = time.Now()
	wait := t.tStart.Sub(t.tCreate)
	ret := t.c.Call(o)
	dura := t.GetRunTime()

	if t.r != nil {
		t.r <- ret
	}

	if t.n != nil {
		SendTaskRes(t.s, t)
	}
	if t.alertTime != 0 && t.name != "" {
		cost := t.GetCostTime()
		if cost > t.alertTime {
			logger.Logger.Warn("task [", t.name, "] since createTime(",
				cost, ") since startTime(", dura, "), in quene wait(", wait, ")")
		}
	}
	return nil
}

func (t *Task) Start() {
	go t.run(nil)
}

func (t *Task) SetAlertTime(alertt time.Duration) {
	t.alertTime = alertt
}

func (t *Task) GetCostTime() time.Duration {
	return time.Now().Sub(t.tCreate)
}

func (t *Task) GetRunTime() time.Duration {
	return time.Now().Sub(t.tStart)
}

func (t *Task) StartByExecutor(name string) bool {
	return sendTaskReqToExecutor(t, name)
}

func (t *Task) StartByFixExecutor(name string) bool {
	return sendTaskReqToFixExecutor(t, name)
}

func (t *Task) BroadcastToAllExecutor() bool {
	return sendTaskReqToAllExecutor(t)
}
