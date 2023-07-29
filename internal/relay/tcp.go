package relay

import (
	"errors"
	"net"
	"time"

	"github.com/mzz2017/softwind/netproxy"
	io2 "github.com/mzz2017/softwind/pkg/zeroalloc/io"
	"github.com/mzz2017/softwind/pool"
	"github.com/mzz2017/softwind/protocol/juicity"
)

func (r *relay) RelayTCP(lConn, rConn netproxy.Conn) (err error) {
	eCh := make(chan error, 1)
	go func() {
		_, e := io2.Copy(rConn, lConn)
		if rConn, ok := rConn.(WriteCloser); ok {
			rConn.CloseWrite()
		}
		rConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		eCh <- e
	}()
	_, e := io2.Copy(lConn, rConn)
	if lConn, ok := lConn.(WriteCloser); ok {
		lConn.CloseWrite()
	}
	lConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if e != nil {
		<-eCh
		return e
	}
	return <-eCh
}

// RelayUoT relays UDP traffict over TCP
func (r *relay) RelayUoT(rDialer netproxy.Dialer, lConn *juicity.PacketConn) (err error) {
	buf := pool.GetFullCap(EthernetMtu)
	defer pool.Put(buf)
	lConn.SetReadDeadline(time.Now().Add(DefaultNatTimeout))
	n, addr, err := lConn.ReadFrom(buf)
	if err != nil {
		return
	}
	conn, err := rDialer.Dial("udp", addr.String())
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil // ignore i/o timeout
		}
		return err
	}
	rConn := conn.(netproxy.PacketConn)
	_ = rConn.SetWriteDeadline(time.Now().Add(DefaultNatTimeout)) // should keep consistent
	_, err = rConn.WriteTo(buf[:n], addr.String())
	if errors.Is(err, net.ErrWriteToConnected) {
		r.logger.Error().Err(err).Msg("relayConnToUDP")
	}
	if err != nil {
		return
	}

	eCh := make(chan error, 1)
	go func() {
		e := r.relayConnToUDP(rConn, lConn, DefaultNatTimeout)
		rConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		eCh <- e
	}()
	e := r.RelayUDPToConn(lConn, rConn, DefaultNatTimeout, len(buf))
	lConn.CloseWrite()
	lConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if e != nil {
		var netErr net.Error
		if errors.As(e, &netErr) && netErr.Timeout() {
			return <-eCh
		}
		<-eCh
		return e
	}
	return <-eCh
}
