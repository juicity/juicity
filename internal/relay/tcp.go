package relay

import (
	"time"

	"github.com/daeuniverse/outbound/netproxy"
	io2 "github.com/daeuniverse/outbound/pkg/zeroalloc/io"
)

func (r *relay) RelayTCP(lConn, rConn netproxy.Conn) (err error) {
	eCh := make(chan error, 1)
	go func() {
		_, e := io2.Copy(rConn, lConn)
		if rConn, ok := rConn.(WriteCloser); ok {
			_ = rConn.CloseWrite()
		}
		_ = rConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		eCh <- e
	}()
	_, e := io2.Copy(lConn, rConn)
	if lConn, ok := lConn.(WriteCloser); ok {
		_ = lConn.CloseWrite()
	}
	_ = lConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if e != nil {
		<-eCh
		return e
	}
	return <-eCh
}
