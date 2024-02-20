package inbound

import (
	"net"
	"net/http"
	"net/netip"

	C "github.com/l552121229/clash-core-backup/constant"
	"github.com/l552121229/clash-core-backup/context"
)

// NewHTTPS receive CONNECT request and return ConnContext
func NewHTTPS(request *http.Request, conn net.Conn) *context.ConnContext {
	metadata := parseHTTPAddr(request)
	metadata.Type = C.HTTPCONNECT
	if ip, port, err := parseAddr(conn.RemoteAddr()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = C.Port(port)
	}
	if addrPort, err := netip.ParseAddrPort(conn.LocalAddr().String()); err == nil {
		metadata.OriginDst = addrPort
	}
	return context.NewConnContext(conn, metadata)
}
