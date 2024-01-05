package peregrine

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/google/uuid"
	"github.com/panjf2000/gnet/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	emptyUpgrader = &ws.Upgrader{}
)

// Conn is upgraded websocket conn
type Conn struct {
	ctx           context.Context
	rwm           sync.RWMutex
	readyUpgraded *atomic.Bool
	LastActive    *atomic.Int64

	// Header all request headers obtained during handshake
	Header http.Header
	ID     string
	Keys   map[string]any

	gnet.Conn
}

func (c *Conn) keepAlive() {
	c.LastActive.Store(time.Now().Unix())
}

func (c *Conn) Context() context.Context {
	return c.ctx
}

func (c *Conn) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *Conn) Set(key string, value any) {
	c.rwm.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
	c.rwm.Unlock()
}

func (c *Conn) Get(key string) (any, bool) {
	c.rwm.RLock()
	value, found := c.Keys[key]
	c.rwm.RUnlock()
	return value, found
}

func NewUpgraderConn(conn gnet.Conn) *Conn {
	lastActive := &atomic.Int64{}
	lastActive.Store(time.Now().Unix())
	return &Conn{
		Conn:          conn,
		LastActive:    lastActive,
		readyUpgraded: &atomic.Bool{},
		ID:            uuid.New().String(),
		ctx:           context.Background(),
	}
}

func TryAssertKeys[T any](c *Conn, key string) (T, bool) {
	val, found := c.Get(key)
	if !found {
		var empty T
		return empty, false
	}
	value, ok := val.(T)
	return value, ok
}
