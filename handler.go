package qWebsocket

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
)

type (
	OnCloseHandlerFunc func(conn *Conn, err error)
	OnPingHandlerFunc  func(conn *Conn)
	HandlerFunc        func(req *HandlerParams)

	HandlerParams struct {
		OpCode  ws.OpCode
		Request []byte
		Writer  io.Writer
		WsConn  *Conn
	}
)

func EmptyOnCloseHandler(_ *Conn, _ error) {}
func EmptyHandler(_ *HandlerParams)        {}

func DefaultOnPingHandler(c *Conn) {
	wsutil.WriteServerMessage(c, ws.OpPong, nil)
}
