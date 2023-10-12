package qWebsocket

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/jellydator/ttlcache/v3"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"
)

type Server struct {
	connNum atomic.Int64
	addr    string

	ctx            context.Context
	engine         gnet.Engine
	workerPool     *ants.Pool
	upgrader       *ws.Upgrader
	keepConnTable  *ttlcache.Cache[string, gnet.Conn]
	logger         Logger
	onCloseHandler OnCloseHandlerFunc
	onPingHandler  OnPingHandlerFunc
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

	if s.logger == nil {
		WithLogger(DefaultLogger)(s)
	}

	if s.onCloseHandler == nil {
		WithOnCloseHandler(EmptyOnCloseHandler)(s)
	}

	if s.onPingHandler == nil {
		WithOnPingHandler(DefaultOnPingHandler)(s)
	}

	if s.handler == nil {
		WithHandler(EmptyHandler)(s)
	}
}

func (s *Server) closeWS(conn *Conn, statusCode ws.StatusCode, reason error) error {
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
		s.connNum.Add(-1)
		upgraderConn, ok := item.Value().Context().(*Conn)
		if !ok {
			_ = item.Value().Close()
			return
		}
		_ = s.closeWS(upgraderConn, ws.StatusGoingAway, errors.New("timeout"))
	})

	// start monitor connect ttl
	go s.keepConnTable.Start()
	go func() {
		time.Sleep(time.Hour * 1)
		s.keepConnTable.DeleteExpired()
	}()
}

func (s *Server) Online() int64 {
	return s.connNum.Load()
}

func (s *Server) CountConnections() int {
	return s.engine.CountConnections()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.engine.Stop(ctx)
}

func (s *Server) ListenAndServe(opts ...gnet.Option) error {
	s.setupTimeoutHandler()
	return gnet.Run(s, s.addr, opts...)
}

// ---- gnet event handler ----

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	s.logger.Infof("[+] Listen addr: %s", s.addr)
	s.engine = eng
	return gnet.None
}

func (s *Server) OnShutdown(_ gnet.Engine) {}

func (s *Server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	s.connNum.Add(1)
	// monitor conn timeout
	s.keepConnTable.Set(c.RemoteAddr().String(), c, ttlcache.DefaultTTL)
	return nil, gnet.None
}

func (s *Server) OnClose(c gnet.Conn, _ error) gnet.Action {
	s.connNum.Add(-1)
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

	upgraderConn, ok := c.Context().(*Conn)
	if !ok {
		s.logger.Errorf("[-] invalid context, remote addr: %s", c.RemoteAddr())
		return gnet.None
	}

	// trying upgrader conn
	if !upgraderConn.successUpgraded.Load() {
		handshake, err := s.upgrader.Upgrade(upgraderConn)
		if err != nil {
			s.logger.Errorf("[-] upgrade error: %s, remote: %s\n", err.Error(), c.RemoteAddr())
			_ = s.closeWS(upgraderConn, ws.StatusProtocolError, err)
			return gnet.Close
		}
		upgraderConn.successUpgraded.Store(true)
		upgraderConn.UpdateActive()
		upgraderConn.Header = handshake.Header
		return gnet.None
	}

	// waiting client message
	messages, err := wsutil.ReadClientMessage(upgraderConn, nil)
	if err != nil {
		s.logger.Errorf("[-] read client message error: %s, remote: %s\n", err.Error(), c.RemoteAddr())
		_ = s.closeWS(upgraderConn, ws.StatusUnsupportedData, err)
		return gnet.Close
	}

	// handle client message
	for _, message := range messages {
		switch message.OpCode {
		case ws.OpPing:
			// async handle
			_ = s.workerPool.Submit(func() {
				s.onPingHandler(upgraderConn)
			})
			upgraderConn.UpdateActive()
		case ws.OpText, ws.OpBinary:
			// async handle
			_ = s.workerPool.Submit(func() {
				s.handler(&HandlerParams{
					OpCode:  message.OpCode,
					Request: message.Payload,
					Writer:  upgraderConn,
					WsConn:  upgraderConn,
				})
			})
			upgraderConn.UpdateActive()
		case ws.OpClose:
			s.onCloseHandler(upgraderConn, nil)
			return gnet.Close
		}

	}
	return gnet.None
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
