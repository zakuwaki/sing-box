package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/option"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"

	"github.com/stretchr/testify/require"
)

func TestECH(t *testing.T) {
	instance := startInstance(t, option.Options{
		DNS: &option.DNSOptions{
			Servers: []option.DNSServerOptions{
				{
					Address: "tls://1.1.1.1",
				},
			},
		},
	})
	dialer, err := tls.NewDialerFromOptions(instance.Router(), N.SystemDialer, "tls-ech.dev", option.OutboundTLSOptions{
		ECH: &option.OutboundECHOptions{
			Enabled: true,
		},
	})
	require.NoError(t, err)
	client := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
			},
			ForceAttemptHTTP2: true,
		},
	}
	response, err := client.Get("https://tls-ech.dev/")
	require.NoError(t, err)
	response.Write(os.Stderr)
}
