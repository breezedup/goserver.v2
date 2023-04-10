package basic

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/breezedup/goserver.v2/core/utils"
)

const (
	DefaultQueueBacklog int = 4
)

var (
// Waitor = utils.NewWaitor()
)

//	 Base class for need alone goroutine objects
//	 that easy to start and when to exit the unified management
//		Feature.
//			establish a tree structure between objects
//			asynchronous message queue
type Object struct {
	*utils.Waitor
	sync.Mutex
	//  Identify
	Id int

	//  Name
	Name string

	//  True if termination was already initiated. If so, we can destroy
	//  the object if there are no more child objects or pending term acks.
	terminating bool

	//  True if termination was already finished.
	terminated bool

	//enlarge que flag
	enlargingQue int32

	//  Sequence number of the last command sent to this object.
	sentSeqnum uint32

	//  Sequence number of the last command processed by this object.
	processedSeqnum uint32

	//  Number of events we have to get before we can destroy the object.
	termAcks int

	//  List of all objects owned by this object. We are responsible
	//  for deallocating them before we quit.
	childs map[int]*Object

	//  Socket owning this object. It's responsible for shutting down
	//  this object.
	owner *Object

	//	Command queue
	que chan Command

	//	Configuration Options
	opt Options

	//	Currently resides goroutine id. I do not know how get it.
	gid int

	//
	waitActive chan struct{}
	//
	waitEnlarge chan struct{}

	//	UserData
	UserData interface{}
	//
	sinker Sinker
	//
	tLastTick time.Time
	//
	timer *time.Ticker
	//object local storage
	ols [OLS_MAX_SLOT]interface{}
}

func NewObject(id int, name string, opt Options, sinker Sinker) *Object {
	o := &Object{
		Id:          id,
		Name:        name,
		opt:         opt,
		sinker:      sinker,
		tLastTick:   time.Now(),
		waitActive:  make(chan struct{}, 1),
		waitEnlarge: make(chan struct{}, 1),
	}

	o.init()
	go o.ProcessCommand()

	return o
}

func (o *Object) init() {
	if o.opt.QueueBacklog < DefaultQueueBacklog {
		o.que = make(chan Command, DefaultQueueBacklog)
	} else {
		o.que = make(chan Command, o.opt.QueueBacklog)
	}
}

// Active inner goroutine
func (o *Object) Active() {
	o.waitActive <- struct{}{}
}

// Launch the supplied object and become its owner.
func (o *Object) LaunchChild(c *Object) {
	if c == nil {
		return
	}

	if c.owner != nil {
		panic("An object can have only one parent node")
	}

	c.owner = o
	c.Waitor = o.Waitor
	c.Active()
	c.safeStart()

	//  Take ownership of the object.
	SendOwn(o, c)
}

//thread safe???
//func (o *Object) GetChildById(id int) *Object {
//	if c, exist := o.childs[id]; exist {
//		return c
//	}
//	return nil
//}

// When another owned object wants to send command to this object
// it calls this function to let it know it should not shut down
// before the command is delivered.
func (o *Object) incSeqnum() {
	atomic.AddUint32(&(o.sentSeqnum), 1)
}

// Special handler called after a command that requires a seqnum
// was processed. The implementation should catch up with its counter
// of processed commands here.
func (o *Object) ProcessSeqnum() {
	//  Catch up with counter of processed commands.
	o.processedSeqnum++

	//  We may have catched up and still have pending terms acks.
	o.checkTermAcks()
}

// Check whether all the peding term acks were delivered.
// If so, deallocate this object.
func (o *Object) checkTermAcks() {
	if o.terminating && o.processedSeqnum == o.sentSeqnum && o.termAcks == 0 {

		//  Sanity check. There should be no active children at this point.

		//  The root object has nobody to confirm the termination to.
		//  Other nodes will confirm the termination to the owner.
		if o.owner != nil {
			SendTermAck(o.owner)
		}

		//  Deallocate the resources.
		o.processDestroy()
	}
}

// Ask owner object to terminate this object. It may take a while
// while actual termination is started. This function should not be
// called more than once.
func (o *Object) Terminate(s *Object) {
	//  If termination is already underway, there's no point
	//  in starting it anew.
	if o.terminating {
		return
	}

	//  As for the root of the ownership tree, there's noone to terminate it,
	//  so it has to terminate itself.
	if o.owner == nil {
		o.processTerm()
		return
	}

	//  If I am an owned object, I'll ask my owner to terminate me.
	SendTermReq(o.owner, o)
}

// Term handler is protocted rather than private so that it can
// be intercepted by the derived class. This is useful to add custom
// steps to the beginning of the termination process.
func (o *Object) processTerm() {
	//  Double termination should never happen.
	if o.terminating {
		return
	}

	//  Send termination request to all owned objects.
	for _, c := range o.childs {
		SendTerm(c)
	}
	o.termAcks += len(o.childs)

	o.safeStop()
	//  Start termination process and check whether by chance we cannot
	//  terminate immediately.
	o.terminating = true
	o.checkTermAcks()
}

// A place to hook in when phyicallal destruction of the object
// is to be delayed.
func (o *Object) processDestroy() {
	o.terminated = true
	close(o.que)
	//clear ols
	o.OlsClrValue()
}

// Enqueue command
func (o *Object) SendCommand(c Command, incseq bool) bool {
	if !atomic.CompareAndSwapInt32(&o.enlargingQue, 0, 0) {
		o.Lock()
		o.Unlock()
	}

	if incseq {
		o.incSeqnum()
	}

	defer func() {
		if err := recover(); err != nil {
			//queue maybe enlarging,and be closed
			o.SendCommand(c, false)
		}
	}()

	//If the queue is full, then enlarge it to two times the size of
redo:
	select {
	case o.que <- c:
	default:
		//Here the lock competition may be more intense when enlarge the beginning,
		//may be enlarge goroutine not to snatch the lock, so in this case is very bad, but no way, let him go
		o.Lock()
		if len(o.que) < cap(o.que) {
			o.Unlock()
			goto redo
		} else {
			if atomic.CompareAndSwapInt32(&o.enlargingQue, 0, 1) {
				defer func() {
					atomic.StoreInt32(&o.enlargingQue, 0)
					o.Unlock()
				}()
				oldCap := cap(o.que)
				newCap := oldCap * 2
				logger.Logger.Warnf("###Object:%v quene cap expand:%v oldcap:%v", o.Name, newCap, oldCap)
				newQue := make(chan Command, newCap)
				//Here closed out queue is to inform other goroutine, then send later.
				close(o.que)
				for cc := range o.que {
					newQue <- cc
				}
				newQue <- c
				o.que = newQue
				return true
			} else {
				o.Unlock()
				runtime.Gosched()
				goto redo
			}
		}
	}

	return true
}

// Dequeue command and process it.
func (o *Object) ProcessCommand() {
	var (
		c        Command
		ok       bool
		tickMode bool
	)

	//wait for active
	<-o.waitActive

	//deamon or no
	if o.Waitor != nil {
		o.Waitor.Add(1)
		defer o.Waitor.Done()
	}

	if o.opt.Interval > 0 && o.sinker != nil && o.timer == nil {
		o.timer = time.NewTicker(o.opt.Interval)
		defer o.timer.Stop()
		tickMode = true
		o.tLastTick = time.Now()
	}

	//There is a small defect;
	//when the queue enlarging, there may be a command sequence can not be guaranteed
	//Because enlarging may occur in other goroutine
	for !o.terminated {
		if !atomic.CompareAndSwapInt32(&o.enlargingQue, 0, 0) {
			//wait enlarge queue
			runtime.Gosched()
			continue
		}
		if tickMode {
			select {
			case c, ok = <-o.que:
				if c != nil {
					o.safeDone(c)
				}
				if !ok {
					if o.terminated {
						return
					} else {
						continue
					}
				}
			case <-o.timer.C:
			}
		} else {
			select {
			case c, ok = <-o.que:
				if c != nil {
					o.safeDone(c)
				}
				if !ok {
					if o.terminated {
						return
					} else {
						continue
					}
				}
			}
		}

		if tickMode && time.Now().After(o.tLastTick.Add(o.opt.Interval-time.Millisecond)) {
			o.safeTick()
			o.tLastTick = time.Now()
			if len(o.que) > o.opt.MaxDone {
				logger.Logger.Warn("(", o.Name, ") object queue cmd count(", len(o.que), ") maxdone(", o.opt.MaxDone, ")")
			}
		}
	}
}

func (o *Object) safeDone(cmd Command) {
	defer utils.DumpStackIfPanic("Object::Command::Done")
	err := cmd.Done(o)
	if err != nil {
		panic(err)
	}
}

func (o *Object) safeStart() {
	defer utils.DumpStackIfPanic("Object::OnStart")

	if o.sinker != nil {
		o.sinker.OnStart()
	}
}

func (o *Object) safeTick() {
	defer utils.DumpStackIfPanic("Object::OnTick")

	if o.sinker != nil {
		o.sinker.OnTick()
	}
}

func (o *Object) safeStop() {
	defer utils.DumpStackIfPanic("Object::OnStop")

	if o.sinker != nil {
		o.sinker.OnStop()
	}
}

func (o *Object) IsTermiated() bool {
	return o.terminated
}
