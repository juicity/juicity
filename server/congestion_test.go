package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"io"
	"math/big"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/protocol/tuic/congestion/bbr"
	"github.com/daeuniverse/quic-go"
	"github.com/daeuniverse/quic-go/congestion"
	"github.com/daeuniverse/quic-go/logging"
	"github.com/rs/zerolog"
)

type observedCongestionConnection struct {
	quic.Connection
	ready chan struct{}
	once  sync.Once
	want  reflect.Type
	bbr   atomic.Bool
}

func (c *observedCongestionConnection) AcceptStream(ctx context.Context) (quic.Stream, error) {
	c.once.Do(func() { close(c.ready) })
	return c.Connection.AcceptStream(ctx)
}
func (c *observedCongestionConnection) SetCongestionControl(cc congestion.CongestionControl) {
	c.Connection.SetCongestionControl(cc)
	c.bbr.Store(reflect.TypeOf(cc) == c.want)
}

func TestCongestionInstalledAtFirstServerDataSend(t *testing.T) {
	for _, hooked := range []bool{true, false} {
		t.Run(map[bool]string{true: "bbr", false: "missing_context"}[hooked], func(t *testing.T) {
			logger := zerolog.Nop()
			s := &Server{logger: &logger, congestionControl: "bbr", cwnd: 10}
			pub, private, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
			der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, private)
			if err != nil {
				t.Fatal(err)
			}
			socket, err := net.ListenPacket("udp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			transport := &quic.Transport{Conn: socket}
			if hooked {
				transport.ConnContext = s.connContext
			}
			defer socket.Close()
			defer transport.Close()
			var observed atomic.Pointer[observedCongestionConnection]
			firstSend := make(chan bool, 1)
			conf := &quic.Config{Tracer: func(ctx context.Context, p logging.Perspective, id quic.ConnectionID) *logging.ConnectionTracer {
				install := s.congestionTracer(ctx, p, id)
				appPending, checked := false, false
				return &logging.ConnectionTracer{
					SentShortHeaderPacket: func(_ *logging.ShortHeader, _ logging.ByteCount, _ logging.ECN, _ *logging.AckFrame, frames []logging.Frame) {
						for _, frame := range frames {
							if stream, ok := frame.(*logging.StreamFrame); ok && stream.Length > 0 {
								appPending = true
							}
						}
					},
					UpdatedMetrics: func(rtt *logging.RTTStats, cwnd, flight logging.ByteCount, packets int) {
						if install != nil {
							install.UpdatedMetrics(rtt, cwnd, flight, packets)
						}
						if appPending && !checked {
							checked = true
							c := observed.Load()
							firstSend <- c != nil && c.bbr.Load()
						}
					},
				}
			}}
			listener, err := transport.Listen(&tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: private}}, NextProtos: []string{"h3"}}, conf)
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			sent := make(chan error, 1)
			handled := make(chan struct{})
			go func() {
				conn, err := listener.Accept(ctx)
				if err != nil {
					sent <- err
					close(handled)
					return
				}
				wrapped := &observedCongestionConnection{Connection: conn, ready: make(chan struct{}), want: reflect.TypeOf(bbr.NewBbrSender(bbr.DefaultClock{}, bbr.GetInitialPacketSize(conn.RemoteAddr())))}
				observed.Store(wrapped)
				go func() { defer close(handled); _ = s.handleConn(wrapped) }()
				select {
				case <-wrapped.ready:
				case <-ctx.Done():
					sent <- ctx.Err()
					return
				}
				stream, err := conn.OpenStreamSync(ctx)
				if err == nil {
					_, err = stream.Write([]byte("server application data"))
					if err == nil {
						err = stream.Close()
					}
				}
				sent <- err
			}()
			client, err := quic.DialAddr(ctx, listener.Addr().String(), &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"h3"}}, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				client.CloseWithError(0, "")
				select {
				case <-handled:
				case <-ctx.Done():
					t.Error("server handler did not stop")
				}
			}()
			stream, err := client.AcceptStream(ctx)
			if err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(stream)
			if err != nil || string(got) != "server application data" {
				t.Fatalf("server data=%q, error=%v", got, err)
			}
			if err := <-sent; err != nil {
				t.Fatal(err)
			}
			select {
			case installed := <-firstSend:
				if installed != hooked {
					t.Fatalf("BBR installed at first application packet=%v, want %v", installed, hooked)
				}
			case <-ctx.Done():
				t.Fatal("first application send was not observed")
			}
		})
	}
}
