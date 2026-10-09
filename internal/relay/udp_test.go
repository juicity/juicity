package relay

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/daeuniverse/quic-go"
	"github.com/rs/zerolog"
)

type frameStream struct {
	quic.Stream
	*bytes.Reader
}

func (s *frameStream) Read(p []byte) (int, error)      { return s.Reader.Read(p) }
func (s *frameStream) SetReadDeadline(time.Time) error { return nil }

type frameDestination struct {
	netproxy.PacketConn
	frames []string
}

func (d *frameDestination) SetWriteDeadline(time.Time) error { return nil }
func (d *frameDestination) WriteTo(p []byte, addr string) (int, error) {
	d.frames = append(d.frames, addr+" "+string(p))
	return len(p), nil
}

func TestUDP443PolicyChecksEachFrame(t *testing.T) {
	var wire bytes.Buffer
	for _, frame := range []struct {
		port uint16
		data string
	}{{53, "first"}, {443, "blocked"}, {5353, "last"}} {
		metadata := trojanc.Metadata{Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: frame.port}, Network: "udp"}
		wire.Write(juicity.SealUDP(metadata, make([]byte, metadata.Len()+2+len(frame.data)), []byte(frame.data)))
	}
	for _, disabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "blocked", false: "allowed"}[disabled], func(t *testing.T) {
			logger := zerolog.Nop()
			r := &relay{logger: &logger}
			src := &juicity.PacketConn{Conn: juicity.NewConn(&frameStream{Reader: bytes.NewReader(wire.Bytes())}, &trojanc.Metadata{
				Metadata: protocol.Metadata{IsClient: true}, Network: "udp",
			}, nil)}
			dst := &frameDestination{}
			err := r.relayConnToUDP(dst, src, time.Second, disabled)
			if !errors.Is(err, io.EOF) {
				t.Fatalf("relay error = %v, want EOF", err)
			}
			want := []string{"127.0.0.1:53 first", "127.0.0.1:5353 last"}
			if !disabled {
				want = []string{"127.0.0.1:53 first", "127.0.0.1:443 blocked", "127.0.0.1:5353 last"}
			}
			if !reflect.DeepEqual(dst.frames, want) {
				t.Fatalf("forwarded frames = %v, want %v", dst.frames, want)
			}
		})
	}
}
