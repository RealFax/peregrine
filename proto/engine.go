package proto

import (
	"bytes"
	qWebsocket "github.com/RealFax/q-websocket"
	"log"
	"os"
	"sync/atomic"
)

type (
	HandlerFunc[T any]                    func(request *Request[T])
	RecoveryFunc[T any, K comparable]     func(protoKey K, request *Request[T], err any)
	NewProtoFunc[T any, K comparable]     func() Proto[T, K]
	DestroyProtoFunc[T any, K comparable] func(params *qWebsocket.HandlerParams, proto Proto[T, K])
)

type Engine[T any, K comparable] struct {
	state atomic.Int32

	codec Codec[T, K]

	handlers map[K]HandlerFunc[T]

	// recovery should be called when the handler panic
	recovery RecoveryFunc[T, K]

	// newProto should be return a Proto interface instance
	newProto NewProtoFunc[T, K]

	// destroyProto should be called after called the handler (if destroyProto not nil)
	destroyProto DestroyProtoFunc[T, K]
}

// handler
//
// impl the handler of q-websocket
func (e *Engine[T, K]) handler(params *qWebsocket.HandlerParams) {
	proto := e.newProto()

	// if registered destroyProto, defer called
	if e.destroyProto != nil {
		defer e.destroyProto(params, proto)
	}

	if err := e.codec.Unmarshal(params.WsConn.ID, bytes.NewReader(params.Request), proto); err != nil {
		log.Println("Proto engine codec error:", err)
		return
	}

	handler, ok := e.handlers[proto.Key()]
	if !ok {
		return
	}

	req := &Request[T]{
		OpCode:     params.OpCode,
		Writer:     params.Writer,
		Conn:       params.WsConn,
		Request:    proto.Self(),
		RawRequest: params.Request,
	}

	// panic handler
	defer func() {
		var err any
		if err = recover(); err == nil {
			return
		}

		if e.recovery != nil {
			e.recovery(proto.Key(), req, err)
			return
		}

		// default recovery handler
		switch tErr := err.(type) {
		case error:
			os.Stderr.WriteString(tErr.Error())
		case string:
			os.Stderr.WriteString(tErr)
		case interface{ String() string }:
			os.Stderr.WriteString(tErr.String())
		}
		os.Exit(0)
	}()

	handler(req)
}

func (e *Engine[T, K]) UseHandler() qWebsocket.HandlerFunc {
	if e.state.Load() == 0 {
		e.state.CompareAndSwap(0, 1)
	}

	return e.handler
}

func (e *Engine[T, K]) Register(key K, handler HandlerFunc[T]) {
	if e.state.Load() == 1 {
		panic("proto: Register should be called before UseHandler")
	}

	e.handlers[key] = handler
}

func (e *Engine[T, K]) RegisterRecovery(recovery RecoveryFunc[T, K]) {
	if e.state.Load() == 1 {
		panic("proto: RegisterRecovery should be called before UseHandler")
	}

	e.recovery = recovery
}

func (e *Engine[T, K]) RegisterCodec(codec Codec[T, K]) {
	if e.state.Load() == 1 {
		panic("proto: RegisterCodec should be called before UseHandler")
	}

	e.codec = codec
}

func (e *Engine[T, K]) RegisterDestroyProto(handler DestroyProtoFunc[T, K]) {
	if e.state.Load() == 1 {
		panic("proto: RegisterDestroyProto should be called before UseHandler")
	}

	e.destroyProto = handler
}
