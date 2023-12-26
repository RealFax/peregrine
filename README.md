# Peregrine
a "high-performance" websocket implementation based on `panjf2000/gnet` and `gobwas/ws`, providing low-level API

# usage
```
go get github.com/RealFax/peregrine@latest
```

## quick start

```go
// echo server

package main

import (
	"github.com/RealFax/peregrine"
	
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/gnet/v2"
	
	"log"
)

func echo(req *peregrine.HandlerParams) {
	wsutil.WriteServerText(req.Writer, req.Request)
}

func main() {
	server := peregrine.NewServer(
		"tcp://0.0.0.0:9090",
		peregrine.WithHandler(echo),
	)

	if err := server.ListenAndServe(gnet.WithMulticore(true)); err != nil {
		log.Fatal("server error:", err)
	}
}
```

## using proto
_proto is a low-level handler built into peregrine, providing simple router_

```go
// using proto 

package main

import (
	"github.com/RealFax/peregrine"
	"github.com/RealFax/peregrine/proto"
	
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	
	"log"
	"strings"
)

// define proto struct

type Proto struct {
	Type    uint32 `json:"type"`
	Message string `json:"message"`
}

func (p *Proto) Key() uint32  { return p.Type }
func (p *Proto) Value() Proto { return *p }
func (p *Proto) Self() *Proto { return p }

const (
	Echo uint32 = iota
	Pong
)

func main() {
	engine := proto.New[Proto, uint32](func() proto.Proto[Proto, uint32] {
		return new(Proto)
	})

	// echo handler
	engine.Register(Echo, func(r *proto.Request[Proto]) {
		r.WriteText([]byte(r.Request.Message))
	})

	// ping pong handler
	engine.Register(Pong, func(r *proto.Request[Proto]) {
		if strings.ToLower(r.Request.Message) == "ping" {
			r.WriteText([]byte("pong"))
			return
		}
		r.WriteClose(ws.StatusGoingAway, "")
	})

	server := peregrine.NewServer(
		"tcp://127.0.0.1:9090",
		peregrine.WithHandler(engine.UseHandler()),
	)

	if err := server.ListenAndServe(gnet.WithMulticore(true)); err != nil {
		log.Fatal("server error:", err)
	}
}
```

more usage see: [Example](https://github.com/RealFax/peregrine/tree/master/example)
