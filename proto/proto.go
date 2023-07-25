package proto

import (
	"bytes"
	qWebsocket "github.com/RealFax/q-websocket"
	"github.com/gobwas/ws"
	"io"
	"log"
)

type Proto[T any, K comparable] interface {
	Key() K
	Value() T
	Self() *T
}

type Codec[T any, K comparable] interface {
	Marshal(w io.Writer, val Proto[T, K]) error
	Unmarshal(r io.Reader, ptr Proto[T, K]) error
}

type Request[T any] struct {
	OpCode     ws.OpCode
	Writer     io.Writer
	Conn       *qWebsocket.GNetUpgraderConn
	Request    T
	RawRequest []byte
}

type HandlerFunc[T any] func(*Request[T])

type Engine[T any, K comparable] struct {
	codec    Codec[T, K]
	handlers map[K]HandlerFunc[T]
	newProto func() Proto[T, K]
}

func (e *Engine[T, K]) Handler(params *qWebsocket.HandlerParams) {
	proto := e.newProto()

	if err := e.codec.Unmarshal(bytes.NewReader(params.Request), proto); err != nil {
		log.Println("Proto engine codec error:", err)
		return
	}

	handler, ok := e.handlers[proto.Key()]
	if !ok {
		return
	}

	handler(&Request[T]{
		OpCode:     params.OpCode,
		Writer:     params.Writer,
		Conn:       params.WsConn,
		Request:    proto.Value(),
		RawRequest: params.Request,
	})
}

func (e *Engine[T, K]) Register(key K, handler HandlerFunc[T]) {
	e.handlers[key] = handler
}

func (e *Engine[T, K]) RegisterCodec(codec Codec[T, K]) {
	e.codec = codec
}

func New[T any, K comparable](newProto func() Proto[T, K]) *Engine[T, K] {
	return &Engine[T, K]{
		codec:    &CodecJSON[T, K]{},
		handlers: make(map[K]HandlerFunc[T]),
		newProto: newProto,
	}
}
