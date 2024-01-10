package peregrine

import (
	"bytes"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type (
	OnCloseHandlerFunc func(conn *Conn, err error)
	OnPingHandlerFunc  func(conn *Conn)
	HandlerFunc        func(packet *Packet)

	Packet struct {
		OpCode  ws.OpCode
		Request []byte
		Conn    *Conn
	}
)

func EmptyHandler(_ *Packet)               {}
func EmptyOnCloseHandler(_ *Conn, _ error) {}
func DefaultOnPingHandler(c *Conn) {
	buf := &bytes.Buffer{}
	_ = wsutil.WriteServerMessage(buf, ws.OpPong, nil)
	_ = c.AsyncWrite(buf.Bytes(), nil)
}
