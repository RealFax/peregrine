package qWebsocket

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/jellydator/ttlcache/v3"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/pkg/errors"
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
	keepConnTable  *ttlcache.Cache[string, gnet.Conn]
	onCloseHandler OnCloseHandlerFunc
	handler        HandlerFunc
}

func (s *Server) withDefault() {
	// check & init

	if s.ctx == nil {
		WithContext(context.Background())(s)
	}

	if s.workerPool == nil {
		WithWorkerPool(1024*1024, ants.Options{
			ExpiryDuration: DefaultWorkerPoolExpiry,
			Nonblocking:    DefaultWorkerPoolNonBlocking,
		})(s)
	}

	if s.upgrader == nil {
		WithUpgrader(emptyUpgrader)(s)
	}

	if s.keepConnTable == nil {
		WithConnTimeout(DefaultConnTimeout)(s)
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

func (s *Server) setupTimeoutHandler() {
	// on connect timeout handler
	s.keepConnTable.OnEviction(func(
		_ context.Context,
		reason ttlcache.EvictionReason,
		item *ttlcache.Item[string, gnet.Conn],
	) {
		atomic.AddInt64(&s.connNum, -1)
		upgraderConn, ok := item.Value().Context().(*GNetUpgraderConn)
		if !ok {
			item.Value().Close()
			return
		}
		s.closeWS(upgraderConn, ws.StatusGoingAway, errors.New("timeout"))
	})

	// start monitor connect ttl
	go s.keepConnTable.Start()
	go func() {
		time.Sleep(time.Hour * 1)
		s.keepConnTable.DeleteExpired()
	}()
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

func (s *Server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	atomic.AddInt64(&s.connNum, 1)
	// monitor conn timeout
	s.keepConnTable.Set(c.RemoteAddr().String(), c, ttlcache.DefaultTTL)
	return nil, gnet.None
}

func (s *Server) OnClose(c gnet.Conn, _ error) gnet.Action {
	atomic.AddInt64(&s.connNum, -1)
	// conn closed, remove conn in monitor list
	s.keepConnTable.Delete(c.RemoteAddr().String())
	return gnet.None
}

func (s *Server) OnTick() (time.Duration, gnet.Action) {
	return 0, gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	// reset conn ttl
	s.keepConnTable.Set(c.RemoteAddr().String(), c, time.Second*5)

	if c.Context() == nil {
		c.SetContext(NewUpgraderConn(c))
	}

	upgraderConn, ok := c.Context().(*GNetUpgraderConn)
	if !ok {
		log.Printf("[-] invalid context, remote addr: %s", c.RemoteAddr())
		return gnet.None
	}

	// trying upgrader conn
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

	// waiting client message
	messages, err := wsutil.ReadClientMessage(upgraderConn, nil)
	if err != nil {
		log.Printf("[-] read client message error: %s, remote: %s\n", err.Error(), c.RemoteAddr())
		s.closeWS(upgraderConn, ws.StatusUnsupportedData, err)
		return gnet.Close
	}

	// handle client message
	for _, message := range messages {
		switch message.OpCode {
		case ws.OpPing:
			wsutil.WriteServerMessage(upgraderConn, ws.OpPong, nil)
			upgraderConn.UpdateActive()
		case ws.OpText, ws.OpBinary:
			// async handle
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
	s.setupTimeoutHandler()
	return gnet.Run(s, s.addr, opts...)
}

func NewServer(addr string, opts ...OptionFunc) *Server {
	s := &Server{
		addr: addr,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.withDefault()
	return s
}
