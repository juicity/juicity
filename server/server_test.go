package server

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"math/big"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/daeuniverse/outbound/protocol/tuic"
	"github.com/daeuniverse/quic-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type countingDialer struct {
	calls atomic.Int32
	dial  func(context.Context, string, string) (netproxy.Conn, error)
}

func (d *countingDialer) DialContext(ctx context.Context, network, addr string) (netproxy.Conn, error) {
	d.calls.Add(1)
	if d.dial != nil {
		return d.dial(ctx, network, addr)
	}
	return nil, errors.New("unexpected dial")
}

func testQUICConnection(t *testing.T, s *Server) quic.Connection {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	socket, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	transport := &quic.Transport{Conn: socket, ConnContext: s.connContext}
	t.Cleanup(func() { _ = transport.Close(); _ = socket.Close() })
	listener, err := transport.Listen(&tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: priv}},
		NextProtos:   []string{"h3"},
	}, &quic.Config{Tracer: s.congestionTracer})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept(ctx)
		if err == nil {
			_ = s.handleConn(conn)
		}
	}()
	conn, err := quic.DialAddr(ctx, listener.Addr().String(), &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"h3"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = conn.CloseWithError(0, "")
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("server handler did not stop")
		}
	})
	return conn
}

func TestIncompleteAuthenticationCannotDial(t *testing.T) {
	old := authenticateTimeout
	authenticateTimeout = 150 * time.Millisecond
	defer func() { authenticateTimeout = old }()
	logger := zerolog.Nop()
	d := &countingDialer{}
	s := &Server{logger: &logger, dialer: d}
	conn := testQUICConnection(t, s)
	auth, err := conn.OpenUniStream()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = auth.Write([]byte{juicity.Version0}); err != nil {
		t.Fatal(err)
	}
	stream, err := conn.OpenStream()
	if err != nil {
		t.Fatal(err)
	}
	request := juicity.NewConn(stream, &trojanc.Metadata{
		Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: 9, IsClient: true},
		Network:  "tcp",
	}, nil)
	if _, err = request.Write(nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-conn.Context().Done():
		var appErr *quic.ApplicationError
		if !errors.As(context.Cause(conn.Context()), &appErr) || appErr.ErrorCode != tuic.AuthenticationFailed {
			t.Fatalf("connection error = %v, want AuthenticationFailed", context.Cause(conn.Context()))
		}
	case <-time.After(time.Second):
		t.Errorf("incomplete authentication connection stayed open past timeout")
	}
	if got := d.calls.Load(); got != 0 {
		t.Errorf("unauthenticated connection dialed %d times", got)
	}
}

func testAuthenticate(t *testing.T, conn quic.Connection, id uuid.UUID) {
	t.Helper()
	token, err := tuic.GenToken(conn.ConnectionState(), id, "password")
	if err != nil {
		t.Fatal(err)
	}
	stream, err := conn.OpenUniStream()
	if err != nil {
		t.Fatal(err)
	}
	writer := bufio.NewWriter(stream)
	if err := tuic.NewAuthenticate(id, token, juicity.Version0).WriteTo(writer); err != nil {
		t.Fatal(err)
	}
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}
}
