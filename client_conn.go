package peregrine

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"sync/atomic"
)

type ClientConn struct {
	state *atomic.Bool
	conn  net.Conn
}

func (c *ClientConn) close(statusCode ws.StatusCode, reason string) error {
	if !c.state.Load() {
		return net.ErrClosed
	}
	c.state.Store(false)
	defer c.conn.Close()
	return ws.WriteFrame(c.conn, ws.NewCloseFrame(ws.NewCloseFrameBody(statusCode, reason)))
}

func (c *ClientConn) Close() error {
	return c.close(ws.StatusNormalClosure, "")
}

func (c *ClientConn) CloseByReason(statusCode ws.StatusCode, reason string) error {
	return c.close(statusCode, reason)
}

func (c *ClientConn) Hijack() (net.Conn, error) {
	if !c.state.Load() {
		return nil, net.ErrClosed
	}
	return c.conn, nil
}

func (c *ClientConn) WriteText(p []byte) error {
	if !c.state.Load() {
		return net.ErrClosed
	}
	return wsutil.WriteClientText(c.conn, p)
}

func (c *ClientConn) WriteBinary(p []byte) error {
	if !c.state.Load() {
		return net.ErrClosed
	}
	return wsutil.WriteClientBinary(c.conn, p)
}

func (c *ClientConn) ReadMessages() ([]wsutil.Message, error) {
	if !c.state.Load() {
		return nil, net.ErrClosed
	}
	return wsutil.ReadServerMessage(c.conn, nil)
}
