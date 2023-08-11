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

// Conn is upgraded websocket conn
type Conn struct {
	gnet.Conn
	LastActive      *atomic.Int64 // atomic
	successUpgraded *atomic.Bool
	ID              string

	ctx context.Context
}

func (c *Conn) Context() context.Context {
	return c.ctx
}

func (c *Conn) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *Conn) UpdateActive() {
	c.LastActive.Store(time.Now().Unix())
}

func NewUpgraderConn(conn gnet.Conn) *Conn {
	lastActive := &atomic.Int64{}
	lastActive.Store(time.Now().Unix())
	return &Conn{
		Conn:            conn,
		LastActive:      lastActive,
		successUpgraded: &atomic.Bool{},
		ID:              uuid.New().String(),
		ctx:             context.Background(),
	}
}
