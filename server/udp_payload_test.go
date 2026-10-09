package server

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/direct"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/google/uuid"
	"github.com/juicity/juicity/internal/relay"
	"github.com/rs/zerolog"
)

func TestUDPStreamPreservesLargePayloads(t *testing.T) {
	sink, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer sink.Close()
	payload := bytes.Repeat([]byte("large UDP payload"), 3500)
	echoDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 65536)
		for range 2 {
			sink.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, addr, err := sink.ReadFromUDP(buf)
			if err != nil {
				echoDone <- err
				return
			}
			_, err = sink.WriteToUDP(buf[:n], addr)
			if err != nil {
				echoDone <- err
				return
			}
		}
		echoDone <- nil
	}()
	logger := zerolog.Nop()
	id := uuid.New()
	s := &Server{logger: &logger, relay: relay.NewRelay(&logger), dialer: direct.FullconeDirect, users: map[uuid.UUID]string{id: "password"}}
	conn := testQUICConnection(t, s)
	testAuthenticate(t, conn, id)
	stream, err := conn.OpenStream()
	if err != nil {
		t.Fatal(err)
	}
	packet := &juicity.PacketConn{Conn: juicity.NewConn(stream, &trojanc.Metadata{Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: uint16(sink.LocalAddr().(*net.UDPAddr).Port), IsClient: true}, Network: "udp"}, nil)}
	buf := make([]byte, 65535)
	for range 2 {
		if _, err := packet.WriteTo(payload, sink.LocalAddr().String()); err != nil {
			t.Fatal(err)
		}
		packet.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _, err := packet.ReadFrom(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf[:n], payload) {
			t.Errorf("round-trip payload = %d bytes, want unchanged %d bytes", n, len(payload))
		}
	}
	if err := <-echoDone; err != nil {
		t.Fatal(err)
	}
}
