package transport

import (
	"context"
	"github.com/pp-chicken/clash-core-backup/common/cache"
	"github.com/pp-chicken/clash-core-backup/constant"
	ClashHttp "github.com/pp-chicken/clash-core-backup/listener/http"
	"net"
	"net/http"
	"net/url"
)

type ClashTransport struct {
	*http.Transport
	in    chan constant.ConnContext
	cache *cache.LruCache
}

func NewClashTransport() *ClashTransport {
	tr := &ClashTransport{
		Transport: &http.Transport{},
		in:        make(chan constant.ConnContext, 100),
		cache:     cache.New(cache.WithAge(30)),
	}
	tr.Transport.Proxy = tr.proxyFunc
	tr.Transport.DialContext = tr.TransportDialContextHandle
	return tr
}

func (c *ClashTransport) GetConnContext() constant.ConnContext {
	return <-c.in
}

func (c *ClashTransport) proxyFunc(*http.Request) (*url.URL, error) {
	return url.Parse("http://127.0.0.1:0")
}

func (c *ClashTransport) TransportDialContextHandle(context.Context, string, string) (net.Conn, error) {
	left, right := net.Pipe()
	go ClashHttp.HandleConn(right, c.in, c.cache)
	return left, nil
}
