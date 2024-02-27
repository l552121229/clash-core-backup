package transport

import (
	"context"
	"github.com/l552121229/clash-core-backup/common/cache"
	"github.com/l552121229/clash-core-backup/constant"
	ClashHttp "github.com/l552121229/clash-core-backup/listener/http"
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
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext:       nil,
		},
		in:    make(chan constant.ConnContext, 100),
		cache: cache.New(cache.WithAge(30)),
	}
	tr.Transport.Proxy = tr.proxyFunc
	tr.Transport.DialContext = tr.TransportDialContextHandle
	return tr
}

func (c *ClashTransport) GetConnContext() constant.ConnContext {
	return <-c.in
}

func (c *ClashTransport) proxyFunc(req *http.Request) (*url.URL, error) {
	return url.Parse("http://127.0.0.1:0")
}

func (c *ClashTransport) TransportDialContextHandle(ctx context.Context, network, addr string) (net.Conn, error) {
	left, right := net.Pipe()

	//启动协程处理连接
	go ClashHttp.HandleConn(right, c.in, c.cache)

	// 返回连接
	return left, nil
}
