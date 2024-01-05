package peregrine

import (
	"fmt"
	"log/slog"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Info(v ...any)
	Infof(format string, v ...any)
	Warnf(format string, args ...interface{})
	Error(v ...any)
	Errorf(format string, v ...any)
	Fatalf(format string, args ...interface{})
}

type logger struct{}

func (l logger) Debugf(format string, args ...interface{}) { slog.Debug(fmt.Sprintf(format, args...)) }
func (l logger) Info(args ...any)                          { slog.Info("", args...) }
func (l logger) Infof(format string, args ...any)          { slog.Info(fmt.Sprintf(format, args...)) }
func (l logger) Warnf(format string, args ...interface{})  { slog.Warn(fmt.Sprintf(format, args...)) }
func (l logger) Error(args ...any)                         { slog.Error("", args...) }
func (l logger) Errorf(format string, args ...any)         { slog.Error(fmt.Sprintf(format, args...)) }
func (l logger) Fatalf(format string, args ...interface{}) { slog.Error(fmt.Sprintf(format, args...)) }

var DefaultLogger = &logger{}
