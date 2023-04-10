package timer

import (
	"github.com/breezedup/goserver.v2/core/basic"
)

type timeoutCommand struct {
	te *TimerEntity
}

func (tc *timeoutCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	if tc.te.stoped {
		return nil
	}
	if tc.te.ta.OnTimer(tc.te.h, tc.te.ud) == false {
		tc.te.stoped = true
		if tc.te.times < 0 {
			StopTimer(tc.te.h)
		}
	}
	return nil
}

func SendTimeout(te *TimerEntity) bool {
	if te.sink == nil {
		return false
	}

	return te.sink.SendCommand(&timeoutCommand{te: te}, true)
}
