package server

import (
	"errors"
	"net"
	"net/netip"
	"time"

	"github.com/miekg/dns"
	"github.com/mzz2017/juice/pkg/log"
	"github.com/mzz2017/softwind/netproxy"
	io2 "github.com/mzz2017/softwind/pkg/zeroalloc/io"
	"github.com/mzz2017/softwind/pool"
	"github.com/mzz2017/softwind/protocol/juice"
)

const (
	EthernetMtu = 1500

	DefaultNatTimeout = 3 * time.Minute
	DnsQueryTimeout   = 17 * time.Second // RFC 5452
)

type WriteCloser interface {
	CloseWrite() error
}

func RelayTCP(lConn, rConn netproxy.Conn) (err error) {
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

func relayConnToUDP(dst netproxy.PacketConn, src *juice.PacketConn, timeout time.Duration) (err error) {
	var n int
	var addr netip.AddrPort
	buf := pool.GetFullCap(EthernetMtu)
	defer pool.Put(buf)
	for {
		_ = src.SetReadDeadline(time.Now().Add(timeout))
		n, addr, err = src.ReadFrom(buf)
		if err != nil {
			return
		}
		_ = dst.SetWriteDeadline(time.Now().Add(DefaultNatTimeout)) // should keep consistent
		_, err = dst.WriteTo(buf[:n], addr.String())
		// WARNING: if the dst is an pre-connected conn, Write should be invoked here.
		if errors.Is(err, net.ErrWriteToConnected) {
			log.Logger().Error().
				Err(err).
				Msg("relayConnToUDP")
		}
		if err != nil {
			return
		}
	}
}

func relayUoT(rDialer netproxy.Dialer, lConn *juice.PacketConn) (err error) {
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
		log.Logger().Error().
			Err(err).
			Msg("relayConnToUDP")
	}
	if err != nil {
		return
	}

	eCh := make(chan error, 1)
	go func() {
		e := relayConnToUDP(rConn, lConn, DefaultNatTimeout)
		rConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		eCh <- e
	}()
	e := RelayUDPToConn(lConn, rConn, DefaultNatTimeout, len(buf))
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

// SelectTimeout selects an appropriate timeout for UDP packet.
func SelectTimeout(packet []byte) time.Duration {
	var dMessage dns.Msg
	if err := dMessage.Unpack(packet); err != nil {
		return DefaultNatTimeout
	}
	return DnsQueryTimeout
}

func RelayUDP(dst *net.UDPConn, laddr net.Addr, src net.PacketConn, timeout time.Duration) (err error) {
	var n int
	buf := pool.GetFullCap(EthernetMtu)
	defer pool.Put(buf)
	for {
		_ = src.SetReadDeadline(time.Now().Add(timeout))
		n, _, err = src.ReadFrom(buf)
		if err != nil {
			return
		}
		_ = dst.SetWriteDeadline(time.Now().Add(DefaultNatTimeout)) // should keep consistent
		_, err = dst.WriteTo(buf[:n], laddr)
		if err != nil {
			return
		}
	}
}

func RelayUDPToConn(dst netproxy.FullConn, src netproxy.PacketConn, timeout time.Duration, bufSize int) (err error) {
	var n int
	var addr netip.AddrPort
	buf := pool.GetFullCap(bufSize)
	defer pool.Put(buf)
	for {
		_ = src.SetReadDeadline(time.Now().Add(timeout))
		n, addr, err = src.ReadFrom(buf)
		if err != nil {
			return
		}
		_ = dst.SetWriteDeadline(time.Now().Add(DefaultNatTimeout)) // should keep consistent
		_, err = dst.WriteTo(buf[:n], addr.String())
		if err != nil {
			return
		}
	}
}
