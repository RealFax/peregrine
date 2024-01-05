package peregrine

import (
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
	wsutil.WriteServerMessage(c, ws.OpPong, nil)
}
