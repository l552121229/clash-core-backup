package main

import (
	"github.com/pp-chicken/clash-core-backup/listener/transport"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	C "github.com/pp-chicken/clash-core-backup/constant"
	"github.com/pp-chicken/clash-core-backup/listener"
	"github.com/pp-chicken/clash-core-backup/tunnel"

	"github.com/stretchr/testify/require"
)

func TestClash_Listener(t *testing.T) {
	basic := `
log-level: silent
port: 7890
socks-port: 7891
redir-port: 7892
tproxy-port: 7893
mixed-port: 7894
`

	err := parseAndApply(basic)
	require.NoError(t, err)
	defer cleanup()

	time.Sleep(waitTime)

	for i := 7890; i <= 7894; i++ {
		require.True(t, TCPing(net.JoinHostPort("127.0.0.1", strconv.Itoa(i))), "tcp port %d", i)
	}
}

func TestClash_ListenerCreate(t *testing.T) {
	basic := `
log-level: silent
`
	err := parseAndApply(basic)
	require.NoError(t, err)
	defer cleanup()

	time.Sleep(waitTime)
	tcpIn := tunnel.TCPIn()
	udpIn := tunnel.UDPIn()

	ports := listener.Ports{
		Port: 7890,
	}
	listener.ReCreatePortsListeners(ports, tcpIn, udpIn)
	require.True(t, TCPing("127.0.0.1:7890"))
	require.Equal(t, ports, *listener.GetPorts())

	inbounds := []C.Inbound{
		{
			Type:        C.InboundTypeHTTP,
			BindAddress: "127.0.0.1:7891",
		},
	}
	listener.ReCreateListeners(inbounds, tcpIn, udpIn)
	require.True(t, TCPing("127.0.0.1:7890"))
	require.Equal(t, ports, *listener.GetPorts())

	require.True(t, TCPing("127.0.0.1:7891"))
	require.Equal(t, len(inbounds), len(listener.GetInbounds()))

	ports.Port = 0
	ports.SocksPort = 7892
	listener.ReCreatePortsListeners(ports, tcpIn, udpIn)
	require.False(t, TCPing("127.0.0.1:7890"))
	require.True(t, TCPing("127.0.0.1:7892"))
	require.Equal(t, ports, *listener.GetPorts())

	require.True(t, TCPing("127.0.0.1:7891"))
	require.Equal(t, len(inbounds), len(listener.GetInbounds()))
}
func TestClash_HttpTransport(t *testing.T) {
	server, err := transport.NewServer(os.Getenv("CLASH_PATH"))
	if err != nil {
		t.Fatal("clash初始化错误", err)
	}
	go server.Run()
	httpClient := http.Client{
		Transport: server.GetTransport(),
		Timeout:   10 * time.Second,
	}
	get, err := httpClient.Get("http://ip-api.com/json")
	if err != nil {
		t.Fatal(err)
	}
	data, readErr := io.ReadAll(get.Body)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if get.StatusCode >= 200 && get.StatusCode < 300 {
		t.Log("请求结果", get.Status, string(data))
		return
	}
	t.Fatal("请求失败", get.Status, string(data))
}
