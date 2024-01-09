package peregrine

import (
	"context"
	"crypto/tls"
	"github.com/gobwas/ws"
	"github.com/jellydator/ttlcache/v3"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"net"
	"time"
)

// ---------- server options ---------- //

type OptionFunc func(*Server)

func WithContext(ctx context.Context) OptionFunc {
	return func(s *Server) { s.ctx = ctx }
}

func WithWorkerPool(size int, options ants.Options) OptionFunc {
	return func(s *Server) { s.workerPool, _ = ants.NewPool(size, ants.WithOptions(options)) }
}

func WithUpgrader(upgrader *ws.Upgrader) OptionFunc {
	return func(s *Server) { s.upgrader = upgrader }
}

func WithConnTimeout(timeout time.Duration) OptionFunc {
	return func(s *Server) {
		s.timeout = timeout
		s.connTable = ttlcache.New[string, gnet.Conn](
			ttlcache.WithTTL[string, gnet.Conn](timeout),
		)
	}
}

func WithLogger(logger Logger) OptionFunc {
	return func(s *Server) { s.logger = logger }
}

func WithOnCloseHandler(handler OnCloseHandlerFunc) OptionFunc {
	return func(s *Server) { s.onCloseHandler = handler }
}

func WithOnPingHandler(handler OnPingHandlerFunc) OptionFunc {
	return func(s *Server) { s.onPingHandler = handler }
}

func WithHandler(handler HandlerFunc) OptionFunc {
	return func(s *Server) { s.handler = handler }
}

// ---------- client options ---------- //

type ClientOptionFunc func(*Client)

func WithClientDialer(dialer ws.Dialer) ClientOptionFunc {
	return func(c *Client) {
		c.dialer = dialer
	}
}

func WithClientNetDialer(
	dialer func(ctx context.Context, network, addr string) (net.Conn, error),
) ClientOptionFunc {
	return func(c *Client) {
		c.dialer.NetDial = dialer
	}
}

func WithClientTLSClient(
	tlsClient func(conn net.Conn, hostname string) net.Conn,
) ClientOptionFunc {
	return func(c *Client) {
		c.dialer.TLSClient = tlsClient
	}
}

func WithClientTLSConfig(tlsConfig *tls.Config) ClientOptionFunc {
	return func(c *Client) {
		c.dialer.TLSConfig = tlsConfig
	}
}

func WithClientWrapConn(wrapConn func(conn net.Conn) net.Conn) ClientOptionFunc {
	return func(c *Client) {
		c.dialer.WrapConn = wrapConn
	}
}
