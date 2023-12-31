package dialer

import (
	"runtime"

	"github.com/daeuniverse/softwind/netproxy"
	"github.com/daeuniverse/softwind/protocol/direct"
	"github.com/juicity/juicity/config"
)

var protectPath string

type clientDialer struct {
	netproxy.Dialer
	conf *config.Config
}

func NewClientDialer(conf *config.Config) *clientDialer {
	return &clientDialer{
		direct.SymmetricDirect,
		conf,
	}
}

func (c *clientDialer) Dial(network string, addr string) (netproxy.Conn, error) {
	if runtime.GOOS == "android" || runtime.GOOS == "linux" {
		protectPath = c.conf.ProtectPath
		if protectPath != "" {
			// Use SoMark func
			magicNetwork := netproxy.MagicNetwork{
				Network: "udp",
				Mark:    114514,
			}
			return c.Dialer.Dial(magicNetwork.Encode(), addr)
		}
	}
	return c.Dialer.Dial(network, addr)
}
