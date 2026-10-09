package server

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type failingPacketConn struct {
	netproxy.PacketConn
	closed  chan struct{}
	written chan string
}

func (c *failingPacketConn) WriteTo(p []byte, addr string) (int, error) {
	c.written <- addr + " " + string(p)
	return 0, errors.New("injected UDP write failure")
}
func (c *failingPacketConn) SetWriteDeadline(time.Time) error { return nil }
func (c *failingPacketConn) Close() error                     { close(c.closed); return nil }

func TestUDPFirstFramePolicyAndWriteFailureCleanup(t *testing.T) {
	for _, blocked := range []bool{true, false} {
		t.Run(map[bool]string{true: "blocked_first_frame_is_skipped", false: "write_failure_closes_socket"}[blocked], func(t *testing.T) {
			blockedFrame := make(chan struct{}, 1)
			logger := zerolog.New(io.Discard).Hook(zerolog.HookFunc(func(_ *zerolog.Event, _ zerolog.Level, message string) {
				if message == "juicity blocked a [udp] frame" {
					blockedFrame <- struct{}{}
				}
			}))
			id := uuid.New()
			outbound := &failingPacketConn{closed: make(chan struct{}), written: make(chan string, 1)}
			d := &countingDialer{dial: func(context.Context, string, string) (netproxy.Conn, error) { return outbound, nil }}
			s := &Server{logger: &logger, dialer: d, users: map[uuid.UUID]string{id: "password"}, disableOutboundUdp443: true}
			conn := testQUICConnection(t, s)
			testAuthenticate(t, conn, id)
			stream, err := conn.OpenStream()
			if err != nil {
				t.Fatal(err)
			}
			port := uint16(53)
			if blocked {
				port = 443
			}
			packet := &juicity.PacketConn{Conn: juicity.NewConn(stream, &trojanc.Metadata{
				Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: port, IsClient: true}, Network: "udp",
			}, nil)}
			if blocked {
				if _, err = packet.WriteTo([]byte("blocked"), "127.0.0.1:443"); err != nil {
					t.Fatal(err)
				}
				select {
				case <-blockedFrame:
				case <-time.After(time.Second):
					t.Fatal("first UDP/443 frame was not skipped")
				}
				if got := d.calls.Load(); got != 0 {
					t.Fatalf("blocked first frame caused %d dials", got)
				}
			}
			if _, err = packet.WriteTo([]byte("allowed"), "127.0.0.1:53"); err != nil {
				t.Fatal(err)
			}
			_ = stream.SetReadDeadline(time.Now().Add(time.Second))
			if _, err = stream.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
				t.Fatalf("stream error = %v, want EOF", err)
			}
			select {
			case got := <-outbound.written:
				if got != "127.0.0.1:53 allowed" {
					t.Fatalf("outbound frame = %q", got)
				}
			case <-time.After(time.Second):
				t.Fatal("allowed frame following blocked first frame was not forwarded")
			}
			select {
			case <-outbound.closed:
			case <-time.After(time.Second):
				t.Fatal("outbound UDP socket leaked after write failure")
			}
		})
	}
}
