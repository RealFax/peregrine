package qWebsocket

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"log"
	"sync/atomic"
	"time"
)

type Server struct {
	connNum int64
	addr    string

	ctx            context.Context
	engine         gnet.Engine
	workerPool     *ants.Pool
	upgrader       *ws.Upgrader
	onCloseHandler OnCloseHandlerFunc
	handler        HandlerFunc
}

func (s *Server) withDefault() {
	if s.ctx == nil {
		WithContext(context.Background())(s)
	}

	if s.workerPool == nil {
		WithWorkerPool(1024*1024, ants.Options{
			ExpiryDuration: time.Second * 10,
			Nonblocking:    true,
		})(s)
	}

	if s.upgrader == nil {
		WithUpgrader(emptyUpgrader)(s)
	}

	if s.onCloseHandler == nil {
		WithOnCloseHandler(EmptyOnCloseHandler)(s)
	}

	if s.handler == nil {
		WithHandler(EmptyHandler)(s)
	}
}

func (s *Server) closeWS(conn *GNetUpgraderConn, statusCode ws.StatusCode, reason error) error {
	defer s.onCloseHandler(conn, reason)
	return ws.WriteFrame(conn, ws.NewCloseFrame(ws.NewCloseFrameBody(statusCode, func() string {
		if reason != nil {
			return reason.Error()
		}
		return ""
	}())))
}

func (s *Server) Online() int64 {
	return atomic.LoadInt64(&s.connNum)
}

func (s *Server) CountConnections() int {
	return s.engine.CountConnections()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.engine.Stop(ctx)
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	log.Printf("[+] Listen addr: %s", s.addr)
	s.engine = eng
	return gnet.None
}

func (s *Server) OnShutdown(_ gnet.Engine) {}

func (s *Server) OnOpen(_ gnet.Conn) ([]byte, gnet.Action) {
	atomic.AddInt64(&s.connNum, 1)
	return nil, gnet.None
}

func (s *Server) OnClose(_ gnet.Conn, _ error) gnet.Action {
	atomic.AddInt64(&s.connNum, -1)
	return gnet.None
}

func (s *Server) OnTick() (time.Duration, gnet.Action) {
	return 0, gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	if c.Context() == nil {
		c.SetContext(NewUpgraderConn(c))
	}

	upgraderConn, ok := c.Context().(*GNetUpgraderConn)
	if !ok {
		log.Printf("[-] invalid context, remote addr: %s", c.RemoteAddr())
		return gnet.None
	}

	if !upgraderConn.successUpgraded {
		if _, err := s.upgrader.Upgrade(upgraderConn); err != nil {
			log.Printf("[-] upgrade error: %s, remote: %s\n", err.Error(), c.RemoteAddr())
			s.closeWS(upgraderConn, ws.StatusProtocolError, err)
			return gnet.Close
		}
		upgraderConn.successUpgraded = true
		upgraderConn.UpdateActive()
		return gnet.None
	}

	messages, err := wsutil.ReadClientMessage(upgraderConn, nil)
	if err != nil {
		s.closeWS(upgraderConn, ws.StatusUnsupportedData, err)
		return gnet.Close
	}

	for _, message := range messages {
		switch message.OpCode {
		case ws.OpPing:
			wsutil.WriteServerMessage(upgraderConn, ws.OpPong, nil)
			upgraderConn.UpdateActive()
		case ws.OpText:
			s.workerPool.Submit(func() {
				s.handler(&HandlerParams{
					OpCode:  message.OpCode,
					Request: message.Payload,
					Writer:  upgraderConn,
					WsConn:  upgraderConn,
				})
			})
			upgraderConn.UpdateActive()
		case ws.OpClose:
			s.closeWS(upgraderConn, ws.StatusNormalClosure, nil)
			return gnet.Close
		}

	}
	return gnet.None
}

func (s *Server) ListenAndServer(opts ...gnet.Option) error {
	return gnet.Run(s, s.addr, opts...)
}

func NewServer(addr string, opts ...OptionFunc) *Server {
	s := &Server{addr: addr}
	for _, opt := range opts {
		opt(s)
	}
	s.withDefault()
	return s
}
