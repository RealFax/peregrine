package peregrine

import (
	"context"
	"github.com/gobwas/ws"
	"io"
	"net/http"
	"sync/atomic"
)

type writer struct {
	n int64
	w io.Writer
}

func (w *writer) WriteString(s string) (int, error) {
	n, err := io.WriteString(w.w, s)
	w.n += int64(n)
	return n, err
}

func (w *writer) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.n += int64(n)
	return n, err
}

type ClientHeader struct {
	http.Header
}

func (h *ClientHeader) WriteTo(w io.Writer) (int64, error) {
	wr := writer{w: w}
	err := h.Write(&wr)
	return wr.n, err
}

type Client struct {
	dialer ws.Dialer
	ctx    context.Context
	addr   string
}

func (c *Client) Dial(header *ClientHeader) (*ClientConn, error) {
	dialer := c.dialer

	// preprocess request header
	if header != nil {
		dialerHeader, ok := dialer.Header.(*ClientHeader)
		if !ok {
			dialerHeader = header
		} else {
			for key, values := range (map[string][]string)(header.Header) {
				dialerHeader.Set(key, values[0])
			}
		}
		dialer.Header = dialerHeader
	}

	conn, _, _, err := dialer.Dial(c.ctx, c.addr)
	if err != nil {
		return nil, err
	}

	return &ClientConn{
		state: func() *atomic.Bool {
			b := atomic.Bool{}
			b.Store(true)
			return &b
		}(),
		conn: conn,
	}, nil
}

func NewClient(addr string, opts ...ClientOptionFunc) *Client {
	client := &Client{
		dialer: ws.DefaultDialer,
		ctx:    context.Background(),
		addr:   addr,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}
