package main

import (
	"testing"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/option"
	N "github.com/sagernet/sing/common/network"
	"github.com/stretchr/testify/require"
	"net/http"
	"net"
	"context"
	M "github.com/sagernet/sing/common/metadata"
	"os"
)

func TestECH(t *testing.T) {
	dialer, err := tls.NewDialerFromOptions(N.SystemDialer, "tls-ech.dev", option.OutboundTLSOptions{
		ECH: &option.OutboundECHOptions{
			Enabled: true,
			Config:  "AEn+DQBFKwAgACABWIHUGj4u+PIggYXcR5JF0gYk3dCRioBW8uJq9H4mKAAIAAEAAQABAANAEnB1YmxpYy50bHMtZWNoLmRldgAA",
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
