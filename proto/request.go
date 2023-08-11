package proto

import (
	"bytes"
	"context"
	qWebsocket "github.com/RealFax/q-websocket"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
)

type Request[T any] struct {
	Context    context.Context
	OpCode     ws.OpCode
	Writer     io.Writer
	Conn       *qWebsocket.Conn
	Request    *T
	RawRequest []byte
}

func (t Request[T]) Reader() io.Reader {
	return bytes.NewReader(t.RawRequest)
}

func (t Request[T]) WriteText(p []byte) error {
	return wsutil.WriteServerMessage(t.Writer, ws.OpText, p)
}

func (t Request[T]) WriteBinary(p []byte) error {
	return wsutil.WriteServerMessage(t.Writer, ws.OpBinary, p)
}

func (t Request[T]) WriteClose(statusCode ws.StatusCode, reason string) error {
	defer t.Conn.Close()
	return wsutil.WriteServerMessage(t.Writer, ws.OpClose, ws.NewCloseFrameBody(statusCode, reason))
}
