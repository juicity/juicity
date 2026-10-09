package relay

import (
	"sync"
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	io2 "github.com/daeuniverse/outbound/pkg/zeroalloc/io"
)

var tcpHalfCloseIdleTimeout = 10 * time.Second

type halfCloseReader struct {
	conn netproxy.Conn
	mu   sync.Mutex
	idle bool
}

func (r *halfCloseReader) Read(p []byte) (int, error) {
	n, err := r.conn.Read(p)
	if n > 0 {
		r.mu.Lock()
		if r.idle {
			_ = r.conn.SetReadDeadline(time.Now().Add(tcpHalfCloseIdleTimeout))
		}
		r.mu.Unlock()
	}
	return n, err
}

func (r *halfCloseReader) startIdleTimeout() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idle = true
	_ = r.conn.SetReadDeadline(time.Now().Add(tcpHalfCloseIdleTimeout))
}

func (r *relay) RelayTCP(lConn, rConn netproxy.Conn) (err error) {
	left := &halfCloseReader{conn: lConn}
	right := &halfCloseReader{conn: rConn}
	eCh := make(chan error, 1)
	go func() {
		_, e := io2.Copy(rConn, left)
		if rConn, ok := rConn.(WriteCloser); ok {
			_ = rConn.CloseWrite()
		}
		right.startIdleTimeout()
		eCh <- e
	}()
	_, e := io2.Copy(lConn, right)
	if lConn, ok := lConn.(WriteCloser); ok {
		_ = lConn.CloseWrite()
	}
	left.startIdleTimeout()
	if e != nil {
		<-eCh
		return e
	}
	return <-eCh
}
