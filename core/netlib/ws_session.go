package netlib

import (
	"bytes"
	"io"
	"strconv"
	"time"

	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/breezedup/goserver.v2/core/utils"
	"github.com/gorilla/websocket"
)

type WsSession struct {
	Session
	conn *websocket.Conn
}

func newWsSession(id int, conn *websocket.Conn, sc *SessionConfig, scl SessionCloseListener) *WsSession {
	s := &WsSession{
		conn: conn,
	}
	s.Session.Id = id
	s.Session.sc = sc
	s.Session.scl = scl
	s.Session.scpl = s
	s.Session.createTime = time.Now()
	s.Session.waitor = utils.NewWaitor()
	s.Session.impl = s
	s.init()

	return s
}

func (s *WsSession) init() {
	s.Session.init()
}

func (s *WsSession) LocalAddr() string {
	return s.conn.LocalAddr().String()
}

func (s *WsSession) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

func (s *WsSession) start() {
	s.lastRcvTime = time.Now()
	go s.sendRoutine()
	go s.recvRoutine()
}

func (s *WsSession) sendRoutine() {
	s.waitor.Add(1)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Trace(s.Id, " ->close: Session.procSend err: ", err)
		}
		ticker.Stop()
		s.sc.encoder.FinishEncode(&s.Session)
		s.shutWrite()
		s.Close()
		s.waitor.Done()
	}()

	b := make([]byte, s.sc.SndBuff)
	buf := bytes.NewBuffer(b)

	var (
		err  error
		data []byte
	)

	for !s.quit || len(s.sendBuffer) != 0 {
		select {
		case packet, ok := <-s.sendBuffer:
			if !ok {
				s.write(websocket.CloseMessage, []byte{})
				panic("[comm expt]sendBuffer chan closed")
			}
			if packet == nil {
				panic("[comm expt]normal close send")
			}
			buf.Reset()
			data, err = s.sc.encoder.Encode(&s.Session, packet.packetid, packet.data, buf)
			if err != nil {
				logger.Logger.Trace("s.sc.encoder.Encode err", err)
				if s.sc.IsInnerLink == false {
					FreePacket(packet)
					panic(err)
				}
			}
			FreePacket(packet)
			if buf.Len() != 0 {
				if s.sc.IsInnerLink {
					var timeZero time.Time
					s.conn.SetWriteDeadline(timeZero)
				} else {
					if s.sc.WriteTimeout != 0 {
						s.conn.SetWriteDeadline(time.Now().Add(s.sc.WriteTimeout))
					}
				}

				if err = s.write(websocket.BinaryMessage, buf.Bytes()); err != nil {
					panic(err)
				}
				s.FirePacketSent(data)
				s.lastSndTime = time.Now()
			}

		case <-ticker.C:
			if err := s.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (s *WsSession) recvRoutine() {
	s.waitor.Add(1)
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Trace(s.Id, " ->close: Session.procRecv err: ", err)
		}
		s.sc.decoder.FinishDecode(&s.Session)
		s.shutRead()
		s.Close()
		s.waitor.Done()
	}()

	s.conn.SetReadLimit(int64(s.sc.MaxPacket))
	var (
		pck      interface{}
		packetid int
		raw      []byte
	)

	for {
		if s.sc.IsInnerLink {
			var timeZero time.Time
			s.conn.SetReadDeadline(timeZero)
		} else {
			if s.sc.ReadTimeout != 0 {
				s.conn.SetReadDeadline(time.Now().Add(s.sc.ReadTimeout))
			}
		}
		op, r, err := s.conn.NextReader()
		if err != nil {
			logger.Logger.Info("s.conn.NextReader err:", err)
			panic(err)
		}
		switch op {
		case websocket.BinaryMessage:
			packetid, pck, err, raw = s.sc.decoder.Decode(&s.Session, r)
			if err != nil {
				bUnproc := true
				bPackErr := false
				if _, ok := err.(*UnparsePacketTypeErr); ok {
					bPackErr = true
					if s.sc.eph != nil && s.sc.eph.OnErrorPacket(&s.Session, packetid, raw) {
						bUnproc = false
					}
				}
				if bUnproc {
					logger.Logger.Warnf("s.sc.decoder.Decode(packetid:%v) err:%v ", packetid, err)
					if s.sc.IsInnerLink == false {
						panic(err)
					} else if !bPackErr {
						panic(err)
					}
				}
			}
			if pck != nil {
				if s.FirePacketReceived(packetid, pck) {
					act := AllocAction()
					act.s = &s.Session
					act.p = pck
					act.packid = packetid
					act.n = "packet:" + strconv.Itoa(packetid)
					s.recvBuffer <- act
				}
			}
			s.lastRcvTime = time.Now()
		}
	}
}

// write writes a message with the given opCode and payload.
func (s *WsSession) write(opCode int, payload []byte) error {
	if !s.sc.IsInnerLink && s.sc.WriteTimeout != 0 {
		s.conn.SetWriteDeadline(time.Now().Add(s.sc.WriteTimeout))
	}
	return s.conn.WriteMessage(opCode, payload)
}

func (s *WsSession) shutRead() {
	if s.shutRecv {
		return
	}
	logger.Logger.Trace(s.Id, " shutRead")
	s.shutRecv = true
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

func (s *WsSession) shutWrite() {
	if s.shutSend {
		return
	}
	logger.Logger.Trace(s.Id, " shutWrite")
	rest := len(s.sendBuffer)
	for rest > 0 {
		packet := <-s.sendBuffer
		if packet != nil {
			FreePacket(packet)
		}
		rest--
	}

	s.shutSend = true
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

func (s *WsSession) onCutPacket(w io.Writer) (err error) {
	if buf, ok := w.(*bytes.Buffer); ok {
		err = s.write(websocket.BinaryMessage, buf.Bytes())
		buf.Reset()
		return
	}
	return
}
