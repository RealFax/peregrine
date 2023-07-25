package qWebsocket

import (
	"github.com/gobwas/ws"
	"io"
)

type OnCloseHandlerFunc func(conn *GNetUpgraderConn, err error)

type HandlerParams struct {
	OpCode  ws.OpCode
	Request []byte
	Writer  io.Writer
	WsConn  *GNetUpgraderConn
}

type HandlerFunc func(req *HandlerParams)

func EmptyOnCloseHandler(_ *GNetUpgraderConn, _ error) {}
func EmptyHandler(_ *HandlerParams)                    {}
