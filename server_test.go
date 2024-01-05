package peregrine_test

import (
	"fmt"
	"github.com/RealFax/peregrine"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/gnet/v2"
	"net/http"
	"net/url"
	"testing"
)

func Handler(req *peregrine.Packet) {
	// echo
	wsutil.WriteServerText(req.Conn, req.Request)
}

func TestServer_ListenAndServer(t *testing.T) {
	server := peregrine.NewServer(
		"tcp://127.0.0.1:9010",
		peregrine.WithUpgrader(&ws.Upgrader{
			OnRequest: peregrine.RequestProxy(func(req *url.URL) error {
				return nil
			}),
			OnHost: peregrine.HostProxy(func(host string) error {
				return nil
			}),
			OnHeader: peregrine.HeaderProxy(func(key, value string) error {
				return nil
			}),
		}),
		peregrine.WithHandler(Handler),
		peregrine.WithOnCloseHandler(func(conn *peregrine.Conn, err error) {
			t.Logf("RemoteAddr: %s close, reason: %v", conn.RemoteAddr(), err)
		}),
	)

	// status monitor
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(fmt.Sprintf("Online: %d", server.ConnTableLen())))
		})
		t.Log("[+] Monitor server: http://localhost:9090")
		http.ListenAndServe("localhost:9090", nil)
	}()

	if err := server.ListenAndServe(
		gnet.WithMulticore(true),
		gnet.WithReuseAddr(true),
		gnet.WithReusePort(true),
	); err != nil {
		t.Fatal(err)
	}
}
