package proto_test

import (
	qWebsocket "github.com/RealFax/q-websocket"
	"github.com/RealFax/q-websocket/proto"
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	"testing"
)

type Proto struct {
	Type    uint32 `json:"type"`
	Message string `json:"msg"`
}

func (p *Proto) Key() uint32  { return p.Type }
func (p *Proto) Value() Proto { return *p }
func (p *Proto) Self() *Proto { return p }

var engine = proto.New[Proto, uint32](func() proto.Proto[Proto, uint32] {
	return new(Proto)
})

func init() {
	engine.Register(1, func(r *proto.Request[Proto]) {
		r.WriteText([]byte(r.Request.Message))
	})
	engine.Register(2, func(r *proto.Request[Proto]) {
		r.WriteClose(ws.StatusGoingAway, "")
	})
}

func TestEngine_Handler(t *testing.T) {
	server := qWebsocket.NewServer(
		"tcp://127.0.0.1:10001",
		qWebsocket.WithHandler(engine.UseHandler()),
	)

	if err := server.ListenAndServe(gnet.WithMulticore(true)); err != nil {
		t.Fatal(err)
	}
}
