package dialer

import (
	"context"
	"runtime"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol/direct"
	"github.com/juicity/juicity/config"
)

var protectPath string

type clientDialer struct {
	Dialer netproxy.Dialer
	conf   *config.Config
}

func NewClientDialer(conf *config.Config) *clientDialer {
	return &clientDialer{
		Dialer: direct.SymmetricDirect,
		conf:   conf,
	}
}

// DialContext implements netproxy.Dialer.
func (c *clientDialer) DialContext(ctx context.Context, network string, addr string) (netproxy.Conn, error) {
	if runtime.GOOS == "android" || runtime.GOOS == "linux" {
		protectPath = c.conf.ProtectPath
		if protectPath != "" {
			// Use SoMark func
			magicNetwork := netproxy.MagicNetwork{
				Network: "udp",
				Mark:    114514,
			}
			return c.Dialer.DialContext(ctx, magicNetwork.Encode(), addr)
		}
	}
	return c.Dialer.DialContext(ctx, network, addr)
}

var _ netproxy.Dialer = (*clientDialer)(nil)
