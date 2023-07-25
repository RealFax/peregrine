package proto_test

import (
	qWebsocket "github.com/RealFax/q-websocket"
	"github.com/RealFax/q-websocket/proto"
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	"log"
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
		log.Println(r.Request.Message)
	})
	engine.Register(2, func(r *proto.Request[Proto]) {
		ws.WriteFrame(r.Writer, ws.NewTextFrame([]byte(r.Request.Message)))
	})
}

func TestEngine_Handler(t *testing.T) {
	server := qWebsocket.NewServer(
		"tcp://127.0.0.1:8080",
		qWebsocket.WithHandler(engine.Handler),
	)

	if err := server.ListenAndServer(gnet.WithMulticore(true)); err != nil {
		t.Fatal(err)
	}
}
