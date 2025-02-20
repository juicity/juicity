package relay

import (
	"net"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol/juicity"

	"github.com/juicity/juicity/pkg/log"
)

type relay struct {
	logger *log.Logger
}

type Relay interface {
	RelayTCP(lConn, rConn netproxy.Conn) (err error)
	RelayUDP(dst net.PacketConn, laddr net.Addr, src net.PacketConn, timeout time.Duration) (err error)
	SelectTimeout(packet []byte) time.Duration
	RelayUoT(rConn netproxy.PacketConn, lConn *juicity.PacketConn, bufLen int) (err error)
	RelayUDPToConn(dst netproxy.FullConn, src netproxy.PacketConn, timeout time.Duration, bufSize int) (err error)
}
type WriteCloser interface {
	CloseWrite() error
}

func NewRelay(logger *log.Logger) Relay {
	return &relay{
		logger: logger,
	}
}
