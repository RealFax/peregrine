package main

import (
	"encoding/json"
	"fmt"
	"github.com/RealFax/peregrine"
	"github.com/RealFax/peregrine/proto"
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ProtoType uint32

const (
	ProtoJoinRoom ProtoType = iota
	ProtoQuitRoom
	ProtoSendMessage

	ProtoRecvMessage = iota - ProtoSendMessage + 1000
)

type Proto struct {
	Type      ProtoType `json:"type"`
	RoomID    uint32    `json:"rid"`
	Timestamp int64     `json:"timestamp,omitempty"`
	Message   string    `json:"msg,omitempty"`
	Signature string    `json:"sign"`
}

func (p *Proto) Key() ProtoType { return p.Type }
func (p *Proto) Value() Proto   { return *p }
func (p *Proto) Self() *Proto   { return p }

type service struct {
	room   sync.Map // map[RoomID]*map[ConnID]struct{}
	online sync.Map // map[ConnID]chan *Proto
}

func (s *service) offline(connID string) {
	s.room.Range(func(_, value any) bool {
		value.(*sync.Map).Delete(connID)
		return false
	})

	ch, ok := s.online.LoadAndDelete(connID)
	if !ok {
		return
	}
	close(ch.(chan *Proto))
}

func (s *service) pushback(roomID uint32, source, msg, sign string) {
	room, ok := s.room.Load(roomID)
	if !ok {
		return
	}

	room.(*sync.Map).Range(func(key, _ any) bool {
		if source == key {
			return false
		}

		ch, cok := s.online.Load(key)
		if !cok {
			return false
		}
		ch.(chan *Proto) <- &Proto{
			Type:      ProtoRecvMessage,
			RoomID:    roomID,
			Timestamp: time.Now().Unix(),
			Message:   msg,
			Signature: sign,
		}
		return false
	})
}

func (s *service) joinRoom(req *proto.Request[Proto]) {
	_, ok := s.online.Load(req.Conn.ID)
	if !ok {
		ch := make(chan *Proto, 4)
		s.online.Store(req.Conn.ID, ch)

		// listen message channel with send
		go func() {
			for {
				msg, ok := <-ch
				if !ok {
					return
				}
				go func() {
					b, _ := json.Marshal(msg)
					ws.WriteFrame(req.Writer, ws.NewTextFrame(b))
				}()
			}
		}()
	}

	room, ok := s.room.Load(req.Request.RoomID)
	if !ok {
		room = &sync.Map{}
		s.room.Store(req.Request.RoomID, room)
	}
	room.(*sync.Map).Store(req.Conn.ID, struct{}{})
}

func (s *service) quitRoom(req *proto.Request[Proto]) {
	room, ok := s.room.Load(req.Request.RoomID)
	if !ok {
		return
	}
	room.(*sync.Map).Delete(req.Conn.ID)
}

func (s *service) sendMessage(req *proto.Request[Proto]) {
	s.pushback(req.Request.RoomID, req.Conn.ID, req.Request.Message, req.Request.Signature)
}

func main() {

	instancePool := proto.NewInstancePool[Proto, ProtoType](func() proto.Proto[Proto, ProtoType] {
		return new(Proto)
	})

	engine := proto.New[Proto, ProtoType](instancePool.Alloc)
	engine.RegisterDestroyProto(func(_ *peregrine.HandlerParams, proto proto.Proto[Proto, ProtoType]) {
		instancePool.Free(proto)
	})

	s := &service{}

	engine.Register(ProtoJoinRoom, s.joinRoom)
	engine.Register(ProtoQuitRoom, s.quitRoom)
	engine.Register(ProtoSendMessage, s.sendMessage)

	server := peregrine.NewServer(
		"tcp://127.0.0.1:8080",
		peregrine.WithHandler(engine.UseHandler()),
		peregrine.WithOnCloseHandler(func(conn *peregrine.Conn, _ error) {
			s.offline(conn.ID)
		}),
	)

	go func() {
		if err := server.ListenAndServe(gnet.WithMulticore(true)); err != nil {
			fmt.Println("[-] gnet error:", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	// server.Stop(context.Background())
}
