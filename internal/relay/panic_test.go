package relay

import (
	"io"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/daeuniverse/quic-go"
	"github.com/rs/zerolog"
)

type panicStream struct {
	quic.Stream
	seen     chan struct{}
	canceled atomic.Bool
}

func (s *panicStream) Read([]byte) (int, error) {
	close(s.seen)
	panic("injected parser panic")
}
func (s *panicStream) SetReadDeadline(time.Time) error  { return nil }
func (s *panicStream) SetWriteDeadline(time.Time) error { return nil }
func (s *panicStream) Close() error                     { return nil }
func (s *panicStream) CancelRead(quic.StreamErrorCode)  { s.canceled.Store(true) }

type panicPeer struct {
	netproxy.PacketConn
	seen <-chan struct{}
}

func (p *panicPeer) SetReadDeadline(time.Time) error { return nil }
func (p *panicPeer) ReadFrom([]byte) (int, netip.AddrPort, error) {
	<-p.seen
	return 0, netip.AddrPort{}, io.EOF
}

func TestRelayGoroutinePanicReturnsErrorAndClosesStream(t *testing.T) {
	logger := zerolog.Nop()
	r := &relay{logger: &logger}
	stream := &panicStream{seen: make(chan struct{})}
	local := &juicity.PacketConn{Conn: juicity.NewConn(stream, &trojanc.Metadata{
		Metadata: protocol.Metadata{IsClient: true}, Network: "udp",
	}, nil)}
	err := r.RelayUoT(&panicPeer{seen: stream.seen}, local, 1500, false)
	if err == nil || !strings.Contains(err.Error(), "UDP stream panic: injected parser panic") {
		t.Fatalf("relay error = %v, want recovered parser panic", err)
	}
	if !stream.canceled.Load() {
		t.Fatal("panicking relay left its stream open")
	}
}
