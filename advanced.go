package qWebsocket

import (
	"github.com/RealFax/q-websocket/common/hack"
	"github.com/gobwas/ws"
	"net/http"
	"net/url"
)

func RequestProxy(proxyFunc func(req *url.URL) error) func(uri []byte) error {
	return func(uri []byte) error {
		if len(uri) == 0 {
			return ws.RejectConnectionError(ws.RejectionStatus(http.StatusBadRequest))
		}

		path, err := url.Parse(hack.Bytes2String(uri))
		if err != nil {
			return ws.RejectConnectionError(
				ws.RejectionStatus(http.StatusBadRequest),
				ws.RejectionReason(err.Error()),
			)
		}

		return proxyFunc(path)
	}
}

func HostProxy(proxyFunc func(host string) error) func(host []byte) error {
	return func(host []byte) error {
		if len(host) == 0 {
			return ws.RejectConnectionError(ws.RejectionStatus(http.StatusBadRequest))
		}
		return proxyFunc(hack.Bytes2String(host))
	}
}

func HeaderProxy(proxyFunc func(key, value string) error) func(key, value []byte) error {
	return func(key, value []byte) error {
		if len(key) == 0 || len(value) == 0 {
			return ws.RejectConnectionError(ws.RejectionStatus(http.StatusBadRequest))
		}
		return proxyFunc(hack.Bytes2String(key), hack.Bytes2String(value))
	}
}
