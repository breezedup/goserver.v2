// session
package netlib

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/breezedup/goserver.v2/core/container"
	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/breezedup/goserver.v2/core/utils"
)

type SessionCloseListener interface {
	onClose(s ISession)
}

type SessionCutPacketListener interface {
	onCutPacket(w io.Writer) error
}

type ISession interface {
	SetAttribute(key, value interface{}) bool
	RemoveAttribute(key interface{})
	GetAttribute(key interface{}) interface{}
	GetSessionConfig() *SessionConfig
	LocalAddr() string
	RemoteAddr() string
	IsIdle() bool
	Close()
	Send(packetid int, data interface{}, sync ...bool) bool
	FireConnectEvent() bool
	FireDisconnectEvent() bool
	FirePacketReceived(packetid int, packet interface{}) bool
	FirePacketSent(data []byte) bool
	FireSessionIdle() bool
}
type packet struct {
	packetid int
	data     interface{}
	next     *packet
}

type Session struct {
	Id          int
	impl        ISession
	sendBuffer  chan *packet
	recvBuffer  chan *action
	sc          *SessionConfig
	attributes  *container.SynchronizedMap
	scl         SessionCloseListener
	scpl        SessionCutPacketListener
	createTime  time.Time
	lastSndTime time.Time
	lastRcvTime time.Time
	waitor      *utils.Waitor
	closed      int32
	quit        bool
	shutSend    bool
	shutRecv    bool
	isConned    bool
}

func (s *Session) init() {
	s.sendBuffer = make(chan *packet, s.sc.MaxPend)
	s.recvBuffer = make(chan *action, s.sc.MaxDone)
	s.attributes = container.NewSynchronizedMap()
}

func (s *Session) SetAttribute(key, value interface{}) bool {
	return s.attributes.Set(key, value)
}

func (s *Session) RemoveAttribute(key interface{}) {
	s.attributes.Delete(key)
}

func (s *Session) GetAttribute(key interface{}) interface{} {
	return s.attributes.Get(key)
}

func (s *Session) GetSessionConfig() *SessionConfig {
	return s.sc
}

func (s *Session) LocalAddr() string {
	if s.impl != nil {
		return s.impl.LocalAddr()
	}
	return ""
}

func (s *Session) RemoteAddr() string {
	if s.impl != nil {
		return s.impl.RemoteAddr()
	}
	return ""
}

func (s *Session) IsConned() bool {
	return s.isConned
}

func (s *Session) IsIdle() bool {
	return s.lastRcvTime.Add(s.sc.IdleTimeout).Before(time.Now())
}

func (s *Session) Close() {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return
	}
	if s.quit {
		return
	}
	s.quit = true

	go s.reapRoutine()
}

func (s *Session) Send(packetid int, data interface{}, sync ...bool) bool {
	if s.quit || s.shutSend {
		return false
	}
	p := AllocPacket()
	p.packetid = packetid
	p.data = data
	if len(sync) > 0 && sync[0] {
		select {
		case s.sendBuffer <- p:
		case <-time.After(time.Duration(s.sc.WriteTimeout)):
			logger.Logger.Warn(s.Id, " send buffer full(", len(s.sendBuffer), "),data be droped(asyn), IsInnerLink ",
				s.sc.IsInnerLink)
			logger.Logger.Warn("Send session(sync) config desc:", *s.sc)
			if s.sc.IsInnerLink == false {
				s.Close()
			}
			return false
		}
	} else {
		select {
		case s.sendBuffer <- p:
		default:
			logger.Logger.Warn(s.Id, " send buffer full(", len(s.sendBuffer), "),data be droped(sync), IsInnerLink ",
				s.sc.IsInnerLink)
			logger.Logger.Warn("Send session(async) config desc:", *s.sc)
			if s.sc.IsInnerLink == false {
				s.Close()
			}
			return false
		}
	}

	return true
}

func (s *Session) FireConnectEvent() bool {
	s.isConned = true
	if s.sc.sfc != nil {
		if !s.sc.sfc.OnSessionOpened(s) {
			return false
		}
	}
	if s.sc.shc != nil {
		s.sc.shc.OnSessionOpened(s)
	}
	return true
}

func (s *Session) FireDisconnectEvent() bool {
	s.isConned = false
	if s.sc.sfc != nil {
		if !s.sc.sfc.OnSessionClosed(s) {
			return false
		}
	}
	if s.sc.shc != nil {
		s.sc.shc.OnSessionClosed(s)
	}
	return true
}

func (s *Session) FirePacketReceived(packetid int, packet interface{}) bool {
	if s.sc.sfc != nil {
		if !s.sc.sfc.OnPacketReceived(s, packetid, packet) {
			return false
		}
	}
	if s.sc.shc != nil {
		s.sc.shc.OnPacketReceived(s, packetid, packet)
	}
	return true
}

func (s *Session) FirePacketSent(data []byte) bool {
	if s.sc.sfc != nil {
		if !s.sc.sfc.OnPacketSent(s, data) {
			return false
		}
	}
	if s.sc.shc != nil {
		s.sc.shc.OnPacketSent(s, data)
	}
	return true
}

func (s *Session) FireSessionIdle() bool {
	if s.sc.sfc != nil {
		if !s.sc.sfc.OnSessionIdle(s) {
			return false
		}
	}
	if s.sc.shc != nil {
		s.sc.shc.OnSessionIdle(s)
	}
	return true
}

func (s *Session) reapRoutine() {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Warn(s.Id, " reapRoutine panic : ", err)
		}
	}()
	if !s.shutSend {
		//close send goroutiue(throw a poison)
		s.sendBuffer <- SendRoutinePoison
	}
	/*
		if !s.shutRecv {
			//close recv goroutiue
			s.shutRead()
		}
	*/
	s.waitor.Wait()
	s.scl.onClose(s)
}

func (s *Session) destroy() {
	s.FireDisconnectEvent()
}
