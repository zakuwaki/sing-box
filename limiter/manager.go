package limiter

import (
	"context"
	"net"

	"github.com/sagernet/sing-box/adapter"
	N "github.com/sagernet/sing/common/network"
)

type Manager interface {
	NewConnWithLimiters(ctx context.Context, conn net.Conn, metadata *adapter.InboundContext, rule adapter.Rule) net.Conn
	NewPacketConnWithLimiters(ctx context.Context, conn N.PacketConn, metadata *adapter.InboundContext, rule adapter.Rule) N.PacketConn
}
