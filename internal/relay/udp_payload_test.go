package relay

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/juicity/juicity/common/consts"
	"github.com/rs/zerolog"
)

func TestUDPFramePayloadBoundary(t *testing.T) {
	payload := bytes.Repeat([]byte{0x5a}, consts.MaxUDPPayload)
	metadata := trojanc.Metadata{Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: 53}, Network: "udp"}
	wire := juicity.SealUDP(metadata, make([]byte, metadata.Len()+2+len(payload)), payload)
	for _, truncated := range []bool{false, true} {
		t.Run(map[bool]string{false: "maximum", true: "truncated"}[truncated], func(t *testing.T) {
			input := wire
			if truncated {
				input = input[:len(input)-1]
			}
			logger := zerolog.Nop()
			src := &juicity.PacketConn{Conn: juicity.NewConn(&frameStream{Reader: bytes.NewReader(input)}, &trojanc.Metadata{Metadata: protocol.Metadata{IsClient: true}, Network: "udp"}, nil)}
			dst := &frameDestination{}
			err := (&relay{logger: &logger}).relayConnToUDP(dst, src, time.Second, false)
			if truncated {
				if !errors.Is(err, io.ErrUnexpectedEOF) || len(dst.frames) != 0 {
					t.Fatalf("truncated frame: error=%v, forwarded=%d", err, len(dst.frames))
				}
			} else if !errors.Is(err, io.EOF) || len(dst.frames) != 1 || dst.frames[0] != "127.0.0.1:53 "+string(payload) {
				t.Fatalf("maximum frame not forwarded intact: error=%v, forwarded=%d", err, len(dst.frames))
			}
		})
	}
}

type oversizedPacketSource struct {
	netproxy.PacketConn
	reads int
}

func (*oversizedPacketSource) SetReadDeadline(time.Time) error { return nil }
func (s *oversizedPacketSource) ReadFrom(p []byte) (int, netip.AddrPort, error) {
	s.reads++
	addr := netip.MustParseAddrPort("127.0.0.1:53")
	switch s.reads {
	case 1:
		return len(p), addr, nil
	case 2:
		return copy(p, "valid"), addr, nil
	default:
		return 0, addr, io.EOF
	}
}

type payloadDestination struct {
	netproxy.FullConn
	payloads []string
}

func (*payloadDestination) SetWriteDeadline(time.Time) error { return nil }
func (d *payloadDestination) WriteTo(p []byte, _ string) (int, error) {
	d.payloads = append(d.payloads, string(p))
	return len(p), nil
}

func TestUDPReturnDropsOversizedPayloadAndContinues(t *testing.T) {
	logger := zerolog.Nop()
	dst := &payloadDestination{}
	err := (&relay{logger: &logger}).RelayUDPToConn(dst, &oversizedPacketSource{}, time.Second)
	if !errors.Is(err, io.EOF) || len(dst.payloads) != 1 || dst.payloads[0] != "valid" {
		t.Fatalf("oversized then valid payload: error=%v, payloads=%v", err, dst.payloads)
	}
}

type oversizedNetSource struct {
	net.PacketConn
	source oversizedPacketSource
}

func (*oversizedNetSource) SetReadDeadline(time.Time) error { return nil }
func (s *oversizedNetSource) ReadFrom(p []byte) (int, net.Addr, error) {
	n, addr, err := s.source.ReadFrom(p)
	return n, net.UDPAddrFromAddrPort(addr), err
}

type netPayloadDestination struct {
	net.PacketConn
	payloads []string
}

func (*netPayloadDestination) SetWriteDeadline(time.Time) error { return nil }
func (d *netPayloadDestination) WriteTo(p []byte, _ net.Addr) (int, error) {
	d.payloads = append(d.payloads, string(p))
	return len(p), nil
}

func TestUDPDropsOversizedPayloadAndContinues(t *testing.T) {
	logger := zerolog.Nop()
	dst := &netPayloadDestination{}
	err := (&relay{logger: &logger}).RelayUDP(dst, &net.UDPAddr{}, &oversizedNetSource{}, time.Second)
	if !errors.Is(err, io.EOF) || len(dst.payloads) != 1 || dst.payloads[0] != "valid" {
		t.Fatalf("oversized then valid payload: error=%v, payloads=%v", err, dst.payloads)
	}
}
