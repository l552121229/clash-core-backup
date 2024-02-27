package transport

import (
	"context"
	"fmt"
	"github.com/pp-chicken/clash-core-backup/adapter/outbound"
	"github.com/pp-chicken/clash-core-backup/config"
	C "github.com/pp-chicken/clash-core-backup/constant"
	"github.com/pp-chicken/clash-core-backup/hub/executor"
	"github.com/pp-chicken/clash-core-backup/log"
	"io"
	"net/http"
)

type Server struct {
	config *config.Config
	trans  *ClashTransport
	direct *outbound.Direct
}

func NewServer(configFilePath string) (*Server, error) {
	C.SetConfig(configFilePath)
	if err := config.Init(C.Path.HomeDir()); err != nil {
		log.Fatalln("Initial configuration directory error: %s", err.Error())
	}
	if configEntity, err := executor.Parse(); err != nil {
		return nil, err
	} else {
		return &Server{
			config: configEntity,
			trans:  NewClashTransport(),
			direct: outbound.NewDirect(),
		}, nil
	}
}

func (s *Server) GetTransport() http.RoundTripper {
	return s.trans
}

func (s *Server) Run() {
	for {
		conn := s.trans.GetConnContext()
		metadata := conn.Metadata()
		go func() {
			for _, v := range s.config.Rules {
				if v.Match(metadata) {
					if provider, exist := s.config.Providers[v.Adapter()]; exist {
						for _, proxyEt := range provider.Proxies() {
							p := proxyEt.Unwrap(metadata)
							if p != nil {
								if proxy(p, conn, metadata) == nil {
									return
								}
							}
						}
					}
				}
			}
			_ = proxy(s.direct, conn, metadata)
		}()
	}
}

func proxy(p C.ProxyAdapter, conn C.ConnContext, metadata *C.Metadata) error {
	remote, err := p.DialContext(context.Background(), metadata)
	if err != nil {
		fmt.Printf("Dial 错误: %s\n", err.Error())
		return err
	}
	r := conn.Conn()
	go func() {
		_, _ = io.Copy(remote, r)
	}()
	_, _ = io.Copy(r, remote)
	return nil
}
