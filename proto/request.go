package proto

import (
	"bytes"
	"context"
	"github.com/RealFax/peregrine"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
)

type Request[T any] struct {
	OpCode ws.OpCode

	Context context.Context
	Conn    *peregrine.Conn
	Request *T
	Payload []byte
}

func (t Request[T]) IsBinary() bool { return t.OpCode == ws.OpBinary }
func (t Request[T]) IsText() bool   { return t.OpCode == ws.OpText }

func (t Request[T]) Reader() io.Reader {
	return bytes.NewReader(t.Payload)
}

func (t Request[T]) WriteText(p []byte) error {
	return wsutil.WriteServerMessage(t.Conn, ws.OpText, p)
}

func (t Request[T]) WriteBinary(p []byte) error {
	return wsutil.WriteServerMessage(t.Conn, ws.OpBinary, p)
}

func (t Request[T]) WriteClose(statusCode ws.StatusCode, reason string) error {
	defer t.Conn.Close()
	return wsutil.WriteServerMessage(t.Conn, ws.OpClose, ws.NewCloseFrameBody(statusCode, reason))
}
