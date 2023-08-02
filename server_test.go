package qWebsocket_test

import (
	"fmt"
	qWebsocket "github.com/RealFax/q-websocket"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/gnet/v2"
	"net/http"
	"net/url"
	"testing"
)

func Handler(req *qWebsocket.HandlerParams) {
	// echo
	wsutil.WriteServerText(req.Writer, req.Request)
}

func TestServer_ListenAndServer(t *testing.T) {
	server := qWebsocket.NewServer(
		"tcp://127.0.0.1:9010",
		qWebsocket.WithUpgrader(&ws.Upgrader{
			OnRequest: qWebsocket.RequestProxy(func(req *url.URL) error {
				return nil
			}),
			OnHost: qWebsocket.HostProxy(func(host string) error {
				return nil
			}),
			OnHeader: qWebsocket.HeaderProxy(func(key, value string) error {
				return nil
			}),
		}),
		qWebsocket.WithHandler(Handler),
	)

	// status monitor
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(fmt.Sprintf("Online: %d", server.Online())))
		})
		t.Log("[+] Monitor server: http://localhost:9090")
		http.ListenAndServe("localhost:9090", nil)
	}()

	if err := server.ListenAndServer(gnet.WithMulticore(true)); err != nil {
		t.Fatal(err)
	}
}
