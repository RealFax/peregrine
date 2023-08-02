package proto

import (
	"bytes"
	qWebsocket "github.com/RealFax/q-websocket"
	"github.com/gobwas/ws"
	"io"
)

type Request[T any] struct {
	OpCode     ws.OpCode
	Writer     io.Writer
	Conn       *qWebsocket.GNetUpgraderConn
	Request    *T
	RawRequest []byte
}

func (t Request[T]) Reader() io.Reader {
	return bytes.NewReader(t.RawRequest)
}

func (t Request[T]) WriteText(p []byte) error {
	return ws.WriteFrame(t.Writer, ws.NewTextFrame(p))
}

func (t Request[T]) WriteBinary(p []byte) error {
	return ws.WriteFrame(t.Writer, ws.NewBinaryFrame(p))
}

func (t Request[T]) WriteClose(statusCode ws.StatusCode, reason string) error {
	return ws.WriteFrame(t.Writer, ws.NewCloseFrame(ws.NewCloseFrameBody(statusCode, reason)))
}
