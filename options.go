package qWebsocket

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/panjf2000/ants/v2"
)

type OptionFunc func(*Server)

func WithContext(ctx context.Context) OptionFunc {
	return func(s *Server) {
		s.ctx = ctx
	}
}

func WithWorkerPool(size int, options ants.Options) OptionFunc {
	return func(s *Server) {
		s.workerPool, _ = ants.NewPool(size, ants.WithOptions(options))
	}
}

func WithUpgrader(upgrader *ws.Upgrader) OptionFunc {
	return func(s *Server) {
		s.upgrader = upgrader
	}
}

func WithOnCloseHandler(handler OnCloseHandlerFunc) OptionFunc {
	return func(s *Server) {
		s.onCloseHandler = handler
	}
}

func WithHandler(handler HandlerFunc) OptionFunc {
	return func(s *Server) {
		s.handler = handler
	}
}
