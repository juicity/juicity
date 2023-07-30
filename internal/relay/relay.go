package relay

import (
	"net"
	"time"

	"github.com/mzz2017/softwind/netproxy"
	"github.com/mzz2017/softwind/protocol/juicity"

	"github.com/juicity/juicity/internal/constant"
	"github.com/juicity/juicity/pkg/log"
)

type relay struct {
	logger          *log.Logger
	mtu             int
	natTimeout      time.Duration
	dnsQueryTimeout time.Duration
}

type Relay interface {
	RelayTCP(lConn, rConn netproxy.Conn) (err error)
	RelayUDP(dst *net.UDPConn, laddr net.Addr, src net.PacketConn, timeout time.Duration) (err error)
	SelectTimeout(packet []byte) time.Duration
	RelayUoT(rDialer netproxy.Dialer, lConn *juicity.PacketConn, fwmark int) (err error)
	RelayUDPToConn(dst netproxy.FullConn, src netproxy.PacketConn, timeout time.Duration, bufSize int) (err error)
}
type WriteCloser interface {
	CloseWrite() error
}

func NewRelay(logger *log.Logger) Relay {
	return &relay{
		logger:          logger,
		mtu:             constant.EthernetMtu,
		natTimeout:      constant.DefaultNatTimeout,
		dnsQueryTimeout: constant.DnsQueryTimeout,
	}
}
