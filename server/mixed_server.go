package server

import (
	"fmt"
	"net"
	"net/url"

	"github.com/mzz2017/softwind/netproxy"
	"github.com/nadoo/glider/proxy"
	"github.com/nadoo/glider/proxy/http"
	"github.com/nadoo/glider/proxy/socks5"
)

type fakeDialer struct {
}

// Addr implements proxy.Dialer.
func (f *fakeDialer) Addr() string {
	return ""
}

// Dial implements proxy.Dialer.
func (f *fakeDialer) Dial(network string, addr string) (c net.Conn, err error) {
	return nil, fmt.Errorf("unimplemented")
}

// DialUDP implements proxy.Dialer.
func (f *fakeDialer) DialUDP(network string, addr string) (pc net.PacketConn, err error) {
	return nil, fmt.Errorf("unimplemented")
}

var defaultFakeDialer proxy.Dialer = &fakeDialer{}

type forwarder struct {
	d netproxy.Dialer
}

// Dial implements proxy.Proxy.
func (f *forwarder) Dial(network string, addr string) (c net.Conn, dialer proxy.Dialer, err error) {
	conn, err := f.d.Dial("tcp", addr)
	if err != nil {
		return nil, defaultFakeDialer, err
	}
	return &netproxy.FakeNetConn{
		Conn:  conn,
		LAddr: nil,
		RAddr: nil,
	}, defaultFakeDialer, nil
}

// DialUDP implements proxy.Proxy.
func (f *forwarder) DialUDP(network string, addr string) (pc net.PacketConn, dialer proxy.UDPDialer, err error) {
	conn, err := f.d.Dial("udp", addr)
	if err != nil {
		return nil, defaultFakeDialer, err
	}
	return &netproxy.FakeNetPacketConn{
		PacketConn: conn.(netproxy.PacketConn),
		LAddr:      nil,
		RAddr:      nil,
	}, defaultFakeDialer, nil
}

// NextDialer implements proxy.Proxy.
func (f *forwarder) NextDialer(dstAddr string) proxy.Dialer {
	return nil
}

// Record implements proxy.Proxy.
func (f *forwarder) Record(dialer proxy.Dialer, success bool) {
}

var _ proxy.Proxy = &forwarder{}

// Mixed struct.
type Mixed struct {
	addr string

	httpServer   *http.HTTP
	socks5Server *socks5.Socks5
}

// NewMixed returns a mixed proxy.
func NewMixed(s string, d netproxy.Dialer) (*Mixed, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("parse err: %s", err)
	}

	p := &forwarder{
		d: d,
	}
	m := &Mixed{
		addr: u.Host,
	}

	m.httpServer, err = http.NewHTTP(s, nil, p)
	if err != nil {
		return nil, err
	}

	m.socks5Server, err = socks5.NewSocks5(s, nil, p)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// ListenAndServe listens on server's addr and serves connections.
func (m *Mixed) ListenAndServe() (err error) {
	go m.socks5Server.ListenAndServeUDP()

	l, err := net.Listen("tcp", m.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", m.addr, err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}

		go m.Serve(c)
	}
}

// Serve serves connections.
func (m *Mixed) Serve(c net.Conn) {
	conn := proxy.NewConn(c)
	if head, err := conn.Peek(1); err == nil {
		if head[0] == socks5.Version {
			m.socks5Server.Serve(c)
			return
		}
	}
	m.httpServer.Serve(conn)
}
