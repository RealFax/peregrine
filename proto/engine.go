package proto

import (
	"bytes"
	"context"
	"github.com/RealFax/peregrine"
	"github.com/gobwas/ws"
	"github.com/pkg/errors"
	"os"
	"runtime/debug"
	"sync/atomic"
)

const (
	KeyErrorCount = "peregrine_proto_error_count"
)

type (
	HandlerFunc[T any]                    func(request *Request[T])
	BrokerFunc[T any]                     func(request *Request[T]) error
	RecoveryFunc[T any, K comparable]     func(protoKey K, request *Request[T], err any)
	NewProtoFunc[T any, K comparable]     func() Proto[T, K]
	DestroyProtoFunc[T any, K comparable] func(params *peregrine.Packet, proto Proto[T, K])
)

type Engine[T any, K comparable] struct {
	Config

	state atomic.Int32

	logger peregrine.Logger

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

func (e *Engine[T, K]) handlerError(packet *peregrine.Packet, err error) {
	e.logger.Errorf("proto: handler error: %s\n", err.Error())

	counter, ready := peregrine.TryAssertKeys[*uint32](packet.Conn, KeyErrorCount)
	if !ready {
		// init ERR_COUNT
		addr := uint32(1)
		packet.Conn.Set(KeyErrorCount, &addr)
		return
	}

	count := atomic.AddUint32(counter, 1)
	if count >= e.MaxErrorCount() {
		defer packet.Conn.Close()
		ws.WriteFrame(
			packet.Conn,
			ws.NewCloseFrame(ws.NewCloseFrameBody(ws.StatusGoingAway, "too many error")),
		)
		return
	}
}

// handler
//
// impl the handler of peregrine
//
// Codec.Unmarshal proto -> find handler -> call brokers -> call handler
func (e *Engine[T, K]) handler(packet *peregrine.Packet) {
	// check request payload size
	if e.MaxPayloadSize() != 0 && uint64(len(packet.Request)) >= e.MaxPayloadSize() {
		e.handlerError(packet, errors.Errorf("request too large, payload size: %d", len(packet.Request)))
		return
	}

	proto := e.newProto()

	// if registered destroyProto, defer called
	if e.destroyProto != nil {
		defer e.destroyProto(packet, proto)
	}

	var err error
	if err = e.codec.Unmarshal(packet.Conn.ID, bytes.NewReader(packet.Request), proto); err != nil {
		e.handlerError(packet, errors.Wrap(err, "codec error"))
		return
	}

	handler, ok := e.handlers[proto.Key()]
	if !ok {
		e.handlerError(packet, errors.Errorf("no %v handler", proto.Key()))
		return
	}

	req := &Request[T]{
		Context: context.Background(),
		OpCode:  packet.OpCode,
		Conn:    packet.Conn,
		Request: proto.Self(),
		Payload: packet.Request,
	}

	// call brokers
	for _, broker := range e.brokers {
		if err = broker(req); err != nil {
			e.handlerError(packet, errors.Wrap(err, "called broker error"))
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

func (e *Engine[T, K]) UseLogger(logger peregrine.Logger) {
	e.logger = logger
}

func (e *Engine[T, K]) UseHandler() peregrine.HandlerFunc {
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
