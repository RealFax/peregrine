package proto

import (
	"bytes"
	"context"
	qWebsocket "github.com/RealFax/q-websocket"
	"github.com/gobwas/ws"
	"github.com/pkg/errors"
	"log"
	"os"
	"runtime/debug"
	"sync/atomic"
)

type (
	HandlerFunc[T any]                    func(request *Request[T])
	BrokerFunc[T any]                     func(request *Request[T]) error
	RecoveryFunc[T any, K comparable]     func(protoKey K, request *Request[T], err any)
	NewProtoFunc[T any, K comparable]     func() Proto[T, K]
	DestroyProtoFunc[T any, K comparable] func(params *qWebsocket.HandlerParams, proto Proto[T, K])
)

type Engine[T any, K comparable] struct {
	state atomic.Int32

	codec Codec[T, K]

	brokers []BrokerFunc[T]

	handlers map[K]HandlerFunc[T]

	// recovery should be called when the handler panic
	recovery RecoveryFunc[T, K]

	// newProto should be return a Proto interface instance
	newProto NewProtoFunc[T, K]

	// destroyProto should be called after called the handler (if destroyProto not nil)
	destroyProto DestroyProtoFunc[T, K]
}

func (e *Engine[T, K]) handlerError(params *qWebsocket.HandlerParams, err error) {
	log.Printf("proto handler error: %s\n", err.Error())

	countAddr, ok := params.WsConn.Context().Value("ERR_COUNT").(*uint32)
	if !ok {
		// init ERR_COUNT
		addr := uint32(1)
		params.WsConn.SetContext(context.WithValue(
			params.WsConn.Context(),
			"ERR_COUNT",
			&addr,
		))
		return
	}

	count := atomic.AddUint32(countAddr, 1)
	if count == 3 {
		defer params.WsConn.Close()
		ws.WriteFrame(
			params.Writer,
			ws.NewCloseFrame(ws.NewCloseFrameBody(ws.StatusGoingAway, "too many error")),
		)
		return
	}
}

// handler
//
// impl the handler of q-websocket
//
// Codec.Unmarshal proto -> find handler -> call brokers -> call handler
func (e *Engine[T, K]) handler(params *qWebsocket.HandlerParams) {
	proto := e.newProto()

	// if registered destroyProto, defer called
	if e.destroyProto != nil {
		defer e.destroyProto(params, proto)
	}

	var err error
	if err = e.codec.Unmarshal(params.WsConn.ID, bytes.NewReader(params.Request), proto); err != nil {
		e.handlerError(params, errors.Wrap(err, "codec error"))
		return
	}

	handler, ok := e.handlers[proto.Key()]
	if !ok {
		e.handlerError(params, errors.Errorf("no %v handler", proto.Key()))
		return
	}

	req := &Request[T]{
		OpCode:     params.OpCode,
		Writer:     params.Writer,
		Conn:       params.WsConn,
		Request:    proto.Self(),
		RawRequest: params.Request,
	}

	// call brokers
	for _, broker := range e.brokers {
		if err = broker(req); err != nil {
			e.handlerError(params, errors.Wrap(err, "called broker error"))
			return
		}
	}

	// panic handler
	defer func() {
		var pErr any
		if pErr = recover(); pErr == nil {
			return
		}

		if e.recovery != nil {
			e.recovery(proto.Key(), req, pErr)
			return
		}

		os.Stderr.WriteString("\nproto panic: ")
		// default recovery handler
		switch val := pErr.(type) {
		case error:
			os.Stderr.WriteString(val.Error())
		case string:
			os.Stderr.WriteString(val)
		case interface{ String() string }:
			os.Stderr.WriteString(val.String())
		}
		os.Stderr.WriteString("\n")
		debug.PrintStack()
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

func (e *Engine[T, K]) UseBrokers(brokers ...BrokerFunc[T]) {
	if e.state.Load() == 1 {
		panic("proto: UseBrokers should be called before UseHandler")
	}
	e.brokers = append(e.brokers, brokers...)
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
