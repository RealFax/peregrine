package peregrine

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/jellydator/ttlcache/v3"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/pkg/errors"
	"time"
)

type Server struct {
	addr string

	ctx        context.Context
	engine     gnet.Engine
	workerPool *ants.Pool
	upgrader   *ws.Upgrader
	connTable  *ttlcache.Cache[string, gnet.Conn]
	logger     Logger

	onCloseHandler OnCloseHandlerFunc
	onPingHandler  OnPingHandlerFunc

	handler HandlerFunc
}

func (s *Server) withDefault() {
	// check & init

	if s.ctx == nil {
		WithContext(context.Background())(s)
	}

	if s.workerPool == nil {
		WithWorkerPool(1024*1024, ants.Options{
			PreAlloc:       true,
			ExpiryDuration: 10 * time.Second,
			Nonblocking:    true,
		})(s)
	}

	if s.upgrader == nil {
		WithUpgrader(emptyUpgrader)(s)
	}

	if s.connTable == nil {
		WithConnTimeout(15 * time.Second)(s)
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

func (s *Server) CloseConn(conn *Conn, statusCode ws.StatusCode, reason error) error {
	defer s.onCloseHandler(conn, reason)
	return ws.WriteFrame(conn, ws.NewCloseFrame(ws.NewCloseFrameBody(statusCode, func() string {
		if reason != nil {
			return reason.Error()
		}
		return ""
	}())))
}

func (s *Server) StartTimeoutScanner() {
	// on connect timeout handler
	s.connTable.OnEviction(func(
		_ context.Context,
		reason ttlcache.EvictionReason,
		item *ttlcache.Item[string, gnet.Conn],
	) {
		upgraderConn, ok := item.Value().Context().(*Conn)
		if !ok {
			_ = item.Value().Close()
			return
		}
		_ = s.CloseConn(upgraderConn, ws.StatusGoingAway, errors.New("timeout"))
	})

	// start monitor connect ttl
	go s.connTable.Start()
	go func() {
		time.Sleep(1 * time.Hour)
		s.connTable.DeleteExpired()
	}()
}

func (s *Server) ConnTableLen() int {
	return s.connTable.Len()
}

func (s *Server) CountConnections() int {
	return s.engine.CountConnections()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.engine.Stop(ctx)
}

func (s *Server) ListenAndServe(opts ...gnet.Option) error {
	s.StartTimeoutScanner()
	return gnet.Run(s, s.addr, append(opts, gnet.WithLogger(s.logger))...)
}

// ---- gnet event handler ----

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	s.logger.Infof("[+] Listen addr: %s", s.addr)
	s.engine = eng
	return gnet.None
}

func (s *Server) OnShutdown(e gnet.Engine) {
	if err := e.Stop(s.ctx); err != nil {
		s.logger.Errorf("gnet.OnShutdown error: %s", err)
	}
}

func (s *Server) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	// monitor conn timeout
	s.connTable.Set(c.RemoteAddr().String(), c, ttlcache.DefaultTTL)
	return nil, gnet.None
}

func (s *Server) OnClose(c gnet.Conn, _ error) gnet.Action {
	// conn closed, remove conn in monitor list
	if addr := c.RemoteAddr(); addr != nil {
		s.connTable.Delete(addr.String())
	}
	return gnet.None
}

func (s *Server) OnTick() (time.Duration, gnet.Action) {
	return 0, gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	// reset conn ttl
	s.connTable.Set(c.RemoteAddr().String(), c, ttlcache.DefaultTTL)

	if c.Context() == nil {
		c.SetContext(NewUpgraderConn(c))
	}

	conn, ok := c.Context().(*Conn)
	if !ok {
		s.logger.Errorf("[-] invalid context, remote addr: %s", c.RemoteAddr())
		return gnet.None
	}

	// trying upgrader conn
	if !conn.readyUpgraded.Load() {
		handshake, err := s.upgrader.Upgrade(conn)
		if err != nil {
			s.logger.Errorf("[-] upgrade error: %s, remote: %s\n", err.Error(), c.RemoteAddr())
			_ = s.CloseConn(conn, ws.StatusProtocolError, err)
			return gnet.Close
		}

		conn.readyUpgraded.Store(true)
		conn.Header = handshake.Header
		conn.keepAlive()
		return gnet.None
	}

	// waiting client message
	messages, err := wsutil.ReadClientMessage(conn, nil)
	if err != nil {
		s.logger.Errorf("[-] read client message error: %s, remote: %s\n", err.Error(), c.RemoteAddr())
		_ = s.CloseConn(conn, ws.StatusUnsupportedData, err)
		return gnet.Close
	}

	// handle client message
	for _, message := range messages {
		switch message.OpCode {
		case ws.OpPing:
			// async handle
			_ = s.workerPool.Submit(func() {
				s.onPingHandler(conn)
			})
			conn.keepAlive()
		case ws.OpText, ws.OpBinary:
			// async handle
			_ = s.workerPool.Submit(func() {
				s.handler(&Packet{
					OpCode:  message.OpCode,
					Request: message.Payload,
					Conn:    conn,
				})
			})
			conn.keepAlive()
		case ws.OpClose:
			s.onCloseHandler(conn, nil)
			return gnet.Close
		default:
			_ = s.CloseConn(conn, ws.StatusUnsupportedData, errors.New("unsupported opcode"))
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
