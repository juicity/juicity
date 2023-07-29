package relay

import (
	"net"
	"time"

	"github.com/mzz2017/softwind/netproxy"
	"github.com/mzz2017/softwind/protocol/juicity"

	"github.com/juicity/juicity/pkg/log"
)

const (
	EthernetMtu = 1500

	DefaultNatTimeout = 3 * time.Minute
	DnsQueryTimeout   = 17 * time.Second // RFC 5452
)

type relay struct {
	logger log.Logger
}

type Relay interface {
	RelayTCP(lConn, rConn netproxy.Conn) (err error)
	RelayUDP(dst *net.UDPConn, laddr net.Addr, src net.PacketConn, timeout time.Duration) (err error)
	SelectTimeout(packet []byte) time.Duration
	RelayUoT(rDialer netproxy.Dialer, lConn *juicity.PacketConn) (err error)
	RelayUDPToConn(dst netproxy.FullConn, src netproxy.PacketConn, timeout time.Duration, bufSize int) (err error)
}
type WriteCloser interface {
	CloseWrite() error
}

func NewRelay() Relay {
	return &relay{logger: log.AccessLogger()}
}
