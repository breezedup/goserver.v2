package task

import (
	"github.com/breezedup/goserver.v2/core/basic"
	"github.com/breezedup/goserver.v2/core/utils"
)

type taskExeCommand struct {
	t *Task
}

func (ttc *taskExeCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	defer utils.DumpStackIfPanic("taskExeCommand")
	return ttc.t.run(o)
}

func SendTaskExe(o *basic.Object, t *Task) bool {
	return o.SendCommand(&taskExeCommand{t: t}, true)
}
