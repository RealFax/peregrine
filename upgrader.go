package qWebsocket

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/google/uuid"
	"github.com/panjf2000/gnet/v2"
	"sync/atomic"
	"time"
)

var (
	emptyUpgrader = &ws.Upgrader{}
)

type GNetUpgraderConn struct {
	gnet.Conn
	LastActive      int64 // atomic
	successUpgraded bool
	ID              string

	ctx context.Context
}

func (c *GNetUpgraderConn) Context() context.Context {
	return c.ctx
}

func (c *GNetUpgraderConn) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *GNetUpgraderConn) UpdateActive() {
	atomic.StoreInt64(&c.LastActive, time.Now().Unix())
}

func NewUpgraderConn(conn gnet.Conn) *GNetUpgraderConn {
	return &GNetUpgraderConn{
		Conn:            conn,
		LastActive:      time.Now().Unix(),
		successUpgraded: false,
		ID:              uuid.New().String(),
	}
}
