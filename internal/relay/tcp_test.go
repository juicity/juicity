package relay

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func tcpPair(t *testing.T) (*net.TCPConn, *net.TCPConn) {
	t.Helper()
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	peer, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	conn, err := listener.AcceptTCP()
	if err != nil {
		peer.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { peer.Close(); conn.Close() })
	return conn, peer
}

func TestTCPHalfCloseIdleTimeout(t *testing.T) {
	old := tcpHalfCloseIdleTimeout
	tcpHalfCloseIdleTimeout = time.Second
	defer func() { tcpHalfCloseIdleTimeout = old }()
	for _, tc := range []struct {
		name   string
		upload bool
		idle   bool
	}{{name: "download"}, {name: "upload", upload: true}, {name: "idle", idle: true}} {
		t.Run(tc.name, func(t *testing.T) {
			left, client := tcpPair(t)
			right, server := tcpPair(t)
			done := make(chan error, 1)
			go func() { done <- (&relay{}).RelayTCP(left, right) }()
			receiver, sender := client, server
			if tc.upload {
				receiver, sender = server, client
			}
			if err := receiver.CloseWrite(); err != nil {
				t.Fatal(err)
			}
			sender.SetReadDeadline(time.Now().Add(10 * time.Second))
			if _, err := io.ReadAll(sender); err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			if !tc.idle {
				writes := make(chan error, 1)
				go func() {
					for range 20 {
						if _, err := sender.Write([]byte("payload")); err != nil {
							writes <- err
							return
						}
						time.Sleep(100 * time.Millisecond)
					}
					writes <- sender.CloseWrite()
				}()
				receiver.SetReadDeadline(time.Now().Add(10 * time.Second))
				got, err := io.ReadAll(receiver)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, bytes.Repeat([]byte("payload"), 20)) {
					t.Errorf("received %d bytes, want full 140-byte response", len(got))
				}
				if err := <-writes; err != nil {
					t.Error(err)
				}
			}
			select {
			case err := <-done:
				if !tc.idle && err != nil {
					t.Fatal(err)
				}
				if tc.idle {
					if e, ok := err.(net.Error); !ok || !e.Timeout() {
						t.Fatalf("idle relay error = %v, want timeout", err)
					}
					if elapsed := time.Since(start); elapsed < tcpHalfCloseIdleTimeout/2 {
						t.Fatalf("idle relay ended too early: %v", elapsed)
					}
				}
			case <-time.After(5 * time.Second):
				t.Fatal("half-closed relay did not finish")
			}
		})
	}
}
