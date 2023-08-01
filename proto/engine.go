package proto

import (
	"bytes"
	qWebsocket "github.com/RealFax/q-websocket"
	"log"
)

type (
	HandlerFunc[T any]                    func(*Request[T])
	NewProtoFunc[T any, K comparable]     func() Proto[T, K]
	DestroyProtoFunc[T any, K comparable] func(params *qWebsocket.HandlerParams, proto Proto[T, K])
)

type Engine[T any, K comparable] struct {
	codec        Codec[T, K]
	handlers     map[K]HandlerFunc[T]
	newProto     NewProtoFunc[T, K]
	destroyProto DestroyProtoFunc[T, K]
}

// Handler
//
// impl the handler of q-websocket
func (e *Engine[T, K]) Handler(params *qWebsocket.HandlerParams) {
	proto := e.newProto()

	if e.destroyProto != nil {
		defer e.destroyProto(params, proto)
	}

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

func (e *Engine[T, K]) RegisterDestroyProto(handler DestroyProtoFunc[T, K]) {
	e.destroyProto = handler
}
