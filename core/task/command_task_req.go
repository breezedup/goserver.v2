package task

import (
	"errors"

	"github.com/breezedup/goserver.v2/core/basic"
	"github.com/breezedup/goserver.v2/core/logger"
)

var (
	TaskErr_CannotFindWorker  = errors.New("Cannot find fit worker.")
	TaskErr_TaskExecuteObject = errors.New("Task can only be executed executor")
)

type taskReqCommand struct {
	t *Task
	n string
}

func (trc *taskReqCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	workerName, err := TaskExecutor.c.Get(trc.n)
	if err != nil {
		logger.Trace("taskReqCommand done error:", err)
		return err
	}
	worker := TaskExecutor.getWorker(workerName)
	if worker != nil {
		logger.Logger.Trace("task[", trc.n, "] dispatch-> worker[", workerName, "]")
		ste := SendTaskExe(worker.Object, trc.t)
		if ste == true {
			logger.Logger.Trace("SendTaskExe success.")
		} else {
			logger.Logger.Trace("SendTaskExe failed.")
		}
		return nil
	} else {
		logger.Logger.Tracef("[%v] worker is no found.", workerName)
		return TaskErr_CannotFindWorker
	}
}

func sendTaskReqToExecutor(t *Task, name string) bool {
	if t == nil {
		logger.Logger.Trace("sendTaskReqToExecutor error,t is nil")
		return false
	}
	if t.n != nil && t.s == nil {
		logger.Logger.Error(name, " You must specify the source object task.")
		return false
	}
	task := TaskExecutor.getWorker(name)
	if task != nil {
		logger.Errorf("[%v] task is not done.", name)
	}
	return TaskExecutor.SendCommand(&taskReqCommand{t: t, n: name}, true)
}

type fixTaskReqCommand struct {
	t *Task
	n string
}

func (trc *fixTaskReqCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	worker := TaskExecutor.getFixWorker(trc.n)
	if worker == nil {
		worker = TaskExecutor.addFixWorker(trc.n)
	}
	if worker != nil {
		logger.Logger.Trace("task[", trc.n, "] dispatch-> worker[", trc.n, "]")
		ste := SendTaskExe(worker.Object, trc.t)
		if ste == true {
			logger.Logger.Trace("SendTaskExe success.")
		} else {
			logger.Logger.Trace("SendTaskExe failed.")
		}
		return nil
	} else {
		logger.Logger.Tracef("[%v] worker is no found.", trc.n)
		return TaskErr_CannotFindWorker
	}
}

func sendTaskReqToFixExecutor(t *Task, name string) bool {
	if t == nil {
		logger.Logger.Trace("sendTaskReqToExecutor error,t is nil")
		return false
	}
	if t.n != nil && t.s == nil {
		logger.Logger.Error(name, " You must specify the source object task.")
		return false
	}
	return TaskExecutor.SendCommand(&fixTaskReqCommand{t: t, n: name}, true)
}

type broadcastTaskReqCommand struct {
	t *Task
}

func (trc *broadcastTaskReqCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	trc.t.AddRefCnt(int32(len(TaskExecutor.workers)))
	//trc.t.AddRefCnt(int32(len(TaskExecutor.fixWorkers)))
	for name, worker := range TaskExecutor.workers {
		//copy
		t := New(trc.t.s, trc.t.c, trc.t.n, trc.t.name+"-"+name)
		//logger.Logger.Trace("task[", t.name, "] dispatch-> worker[", name, "]")
		SendTaskExe(worker.Object, t)
	}
	//	for name, worker := range TaskExecutor.fixWorkers {
	//		//copy
	//		t := New(trc.t.s, trc.t.c, trc.t.n, trc.t.name+"-"+name)
	//		//logger.Logger.Trace("task[", t.name, "] dispatch-> worker[", name, "]")
	//		SendTaskExe(worker.Object, t)
	//	}
	return nil
}

func sendTaskReqToAllExecutor(t *Task) bool {
	if t == nil {
		logger.Logger.Trace("sendTaskReqToExecutor error,t is nil")
		return false
	}
	return TaskExecutor.SendCommand(&broadcastTaskReqCommand{t: t}, true)
}
