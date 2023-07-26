package qWebsocket

import (
	"context"
	"crypto/tls"
	"github.com/gobwas/ws"
	"github.com/panjf2000/ants/v2"
	"net"
)

// ---------- server options ---------- //

type OptionFunc func(*Server)

func WithContext(ctx context.Context) OptionFunc {
	return func(s *Server) {
		s.ctx = ctx
	}
}

func WithWorkerPool(size int, options ants.Options) OptionFunc {
	return func(s *Server) {
		s.workerPool, _ = ants.NewPool(size, ants.WithOptions(options))
	}
}

func WithUpgrader(upgrader *ws.Upgrader) OptionFunc {
	return func(s *Server) {
		s.upgrader = upgrader
	}
}

func WithOnCloseHandler(handler OnCloseHandlerFunc) OptionFunc {
	return func(s *Server) {
		s.onCloseHandler = handler
	}
}

func WithHandler(handler HandlerFunc) OptionFunc {
	return func(s *Server) {
		s.handler = handler
	}
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
