package server

import (
	"bytes"
	"io"
	"net/netip"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/juicity/juicity/common/consts"
	"github.com/rs/zerolog"
)

type endpointPacketSource struct {
	netproxy.PacketConn
	payload []byte
	reads   int
}

func (s *endpointPacketSource) ReadFrom(p []byte) (int, netip.AddrPort, error) {
	s.reads++
	addr := netip.MustParseAddrPort("127.0.0.1:53")
	switch s.reads {
	case 1:
		return consts.MaxUDPPayload + 1, addr, nil
	case 2:
		return copy(p, s.payload), addr, nil
	default:
		return 0, addr, io.EOF
	}
}
func TestUDPEndpointDropsOversizeAndPreservesLargePayload(t *testing.T) {
	logger := zerolog.Nop()
	payload := bytes.Repeat([]byte{0x4b}, 59500)
	var received [][]byte
	endpoint := &UdpEndpoint{logger: &logger, conn: &endpointPacketSource{payload: payload}, NatTimeout: time.Minute, deadlineTimer: time.NewTimer(time.Minute), handler: func(p []byte, _ netip.AddrPort, _ any) error { received = append(received, bytes.Clone(p)); return nil }}
	endpoint.start()
	if len(received) != 1 || !bytes.Equal(received[0], payload) {
		t.Fatalf("endpoint forwarded %d payloads, want one intact 59500-byte payload", len(received))
	}
}
