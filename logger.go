package qWebsocket

import (
	"fmt"
	"log/slog"
)

type Logger interface {
	Info(v ...any)
	Infof(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
}

type logger struct{}

func (l logger) Info(args ...any)               { slog.Info("", args...) }
func (l logger) Infof(msg string, args ...any)  { slog.Info(fmt.Sprintf(msg, args...)) }
func (l logger) Error(args ...any)              { slog.Error("", args...) }
func (l logger) Errorf(msg string, args ...any) { slog.Error(fmt.Sprintf(msg, args...)) }

var DefaultLogger = &logger{}
