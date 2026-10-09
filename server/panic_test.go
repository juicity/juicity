package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestStreamPanicDoesNotCloseAuthenticatedConnection(t *testing.T) {
	old := authenticateTimeout
	authenticateTimeout = 150 * time.Millisecond
	defer func() { authenticateTimeout = old }()
	logger := zerolog.Nop()
	id := uuid.New()
	d := &countingDialer{}
	d.dial = func(context.Context, string, string) (netproxy.Conn, error) {
		if d.calls.Load() == 1 {
			panic("injected stream panic")
		}
		return nil, errors.New("dial complete")
	}
	s := &Server{logger: &logger, dialer: d, users: map[uuid.UUID]string{id: "password"}}
	conn := testQUICConnection(t, s)
	testAuthenticate(t, conn, id)
	for n := int32(1); n <= 2; n++ {
		stream, err := conn.OpenStream()
		if err != nil {
			t.Fatal(err)
		}
		request := juicity.NewConn(stream, &trojanc.Metadata{
			Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: 9, IsClient: true}, Network: "tcp",
		}, nil)
		if _, err = request.Write(nil); err != nil {
			t.Fatal(err)
		}
		_ = stream.SetReadDeadline(time.Now().Add(time.Second))
		_, err = stream.Read(make([]byte, 1))
		var timeout interface{ Timeout() bool }
		if err == nil || (errors.As(err, &timeout) && timeout.Timeout()) {
			t.Fatalf("stream did not close: %v", err)
		}
		if got := d.calls.Load(); got != n {
			t.Fatalf("dial count = %d, want %d", got, n)
		}
		if n == 1 {
			// Successful authentication must survive the original authentication deadline.
			select {
			case <-conn.Context().Done():
				t.Fatalf("stream panic closed the connection: %v", context.Cause(conn.Context()))
			case <-time.After(2 * authenticateTimeout):
			}
		}
	}
}
