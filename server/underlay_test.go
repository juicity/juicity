package server

import (
	"bytes"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/ciphers"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/direct"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/shadowsocks"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/rs/zerolog"
)

func TestUnderlayFirstPacketAuthenticatesAndForwardsPlaintext(t *testing.T) {
	sink, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer sink.Close()
	_, portString, _ := net.SplitHostPort(sink.LocalAddr().String())
	port, _ := strconv.Atoi(portString)
	logger := zerolog.Nop()
	s := &Server{logger: &logger, dialer: direct.FullconeDirect,
		inFlightUnderlayKey: NewInFlightUnderlayKey(time.Second), udpEndpointPool: NewUdpEndpointPool()}
	key := &shadowsocks.Key{CipherConf: juicity.CipherConf, MasterKey: make([]byte, juicity.CipherConf.KeyLen)}
	payload := []byte("authenticated underlay payload")
	salt := make([]byte, juicity.CipherConf.SaltLen)
	salt[2] = 1
	s.inFlightUnderlayKey.Store(inFlightKey(salt), &juicity.UnderlayAuth{
		IV: salt, Psk: key.MasterKey,
		Metadata: &trojanc.Metadata{Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: uint16(port)}, Network: "udp"},
	})
	waiting := newUnderlayWaitBudget()
	for _, tampered := range []bool{true, false} {
		packet, err := shadowsocks.EncryptUDPFromPool(key, payload, salt, ciphers.JuicityReusedInfo)
		if err != nil {
			t.Fatal(err)
		}
		if tampered {
			packet[len(packet)-1] ^= 1
		}
		source := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
		err = s.dispatchNonQuicPacket(nil, packet, source, waiting)
		packet.Put()
		if tampered {
			if err == nil {
				t.Fatal("tampered first packet accepted")
			}
			if _, ok := s.udpEndpointPool.pool.Load(source.AddrPort()); ok {
				t.Fatal("tampered first packet created an outbound endpoint")
			}
			if s.inFlightUnderlayKey.Lookup(inFlightKey(salt)) == nil {
				t.Fatal("tampered first packet consumed the credential")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		_ = sink.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 1500)
		n, _, err := sink.ReadFrom(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf[:n], payload) {
			t.Fatalf("first packet payload = %x, want %x", buf[:n], payload)
		}
		endpoint, _ := s.udpEndpointPool.pool.Load(source.AddrPort())
		defer s.udpEndpointPool.Remove(source.AddrPort(), endpoint.(*UdpEndpoint))
	}
}
