package server

import (
	"bytes"
	"context"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/ciphers"
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/direct"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/shadowsocks"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/rs/zerolog"
)

func TestUnderlayWaitBudgetDoesNotDropLiveEndpoint(t *testing.T) {
	for _, global := range []bool{false, true} {
		t.Run(map[bool]string{false: "per_source_IP", true: "global"}[global], func(t *testing.T) {
			sink, err := net.ListenPacket("udp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer sink.Close()
			logger := zerolog.Nop()
			s := &Server{logger: &logger, dialer: direct.FullconeDirect,
				inFlightUnderlayKey: NewInFlightUnderlayKey(3 * time.Second), udpEndpointPool: NewUdpEndpointPool()}
			defer func() {
				s.inFlightUnderlayKey.mu.Lock()
				defer s.inFlightUnderlayKey.mu.Unlock()
				for _, cc := range s.inFlightUnderlayKey.notify {
					cc.Cancel()
				}
			}()
			key := &shadowsocks.Key{CipherConf: juicity.CipherConf, MasterKey: make([]byte, juicity.CipherConf.KeyLen)}
			source := &net.UDPAddr{IP: net.IPv4(127, 1, 0, 1), Port: 9999}
			endpoint, _, err := s.udpEndpointPool.GetOrCreate(source.AddrPort(), &UdpEndpointOptions{
				NatTimeout: 10 * time.Second,
				Handler:    func([]byte, netip.AddrPort, any) error { return nil },
				GetDialOption: func() (*DialOption, error) {
					return &DialOption{Target: sink.LocalAddr().String(), Dialer: direct.FullconeDirect, Metadata: key.MasterKey}, nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer s.udpEndpointPool.Remove(source.AddrPort(), endpoint)
			waiting := newUnderlayWaitBudget()
			limit := maxPendingUnderlayPerIP
			if global {
				limit = maxPendingUnderlayHandlers
			}
			for n := range limit + 1 {
				unknown := make([]byte, juicity.CipherConf.SaltLen)
				unknown[2], unknown[3], unknown[4] = 200, byte(n), byte(n>>8)
				ip := net.IPv4(127, 1, 0, 1)
				if global {
					ip = net.IPv4(127, 1, byte(n/maxPendingUnderlayPerIP), 1)
				}
				if err := s.dispatchNonQuicPacket(nil, unknown, &net.UDPAddr{IP: ip, Port: 10000 + n}, waiting); err != nil {
					t.Fatal(err)
				}
			}
			if got := len(waiting.pending); got != limit {
				t.Fatalf("pending waits = %d, want cap %d", got, limit)
			}
			payload := []byte("live endpoint survives exhausted credential budget")
			packet, err := shadowsocks.EncryptUDPFromPool(key, payload, make([]byte, juicity.CipherConf.SaltLen), ciphers.JuicityReusedInfo)
			if err != nil {
				t.Fatal(err)
			}
			err = s.dispatchNonQuicPacket(nil, packet, source, waiting)
			packet.Put()
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
				t.Fatalf("live endpoint payload = %x, want %x", buf[:n], payload)
			}
			salt := make([]byte, juicity.CipherConf.SaltLen)
			salt[2] = 42
			metadata, err := protocol.ParseMetadata(sink.LocalAddr().String())
			if err != nil {
				t.Fatal(err)
			}
			s.inFlightUnderlayKey.Store(inFlightKey(salt), &juicity.UnderlayAuth{
				Psk: key.MasterKey, Metadata: &trojanc.Metadata{Metadata: metadata, Network: "udp"},
			})
			firstSource := &net.UDPAddr{IP: source.IP, Port: 9998}
			firstPayload := []byte("authenticated first packet bypasses exhausted waits")
			packet, err = shadowsocks.EncryptUDPFromPool(key, firstPayload, salt, ciphers.JuicityReusedInfo)
			if err != nil {
				t.Fatal(err)
			}
			err = s.dispatchNonQuicPacket(nil, packet, firstSource, waiting)
			packet.Put()
			if err != nil {
				t.Fatal(err)
			}
			n, _, err = sink.ReadFrom(buf)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf[:n], firstPayload) {
				t.Fatalf("authenticated first payload = %x, want %x", buf[:n], firstPayload)
			}
			firstEndpoint, _ := s.udpEndpointPool.pool.Load(firstSource.AddrPort())
			defer s.udpEndpointPool.Remove(firstSource.AddrPort(), firstEndpoint.(*UdpEndpoint))
		})
	}
}

func TestUnderlayHandlerPanicDropsPacket(t *testing.T) {
	panicked := make(chan struct{}, 1)
	logger := zerolog.New(&bytes.Buffer{}).Hook(zerolog.HookFunc(func(_ *zerolog.Event, _ zerolog.Level, message string) {
		if message == "underlay handler panic: packet dropped" {
			panicked <- struct{}{}
		}
	}))
	d := &countingDialer{dial: func(context.Context, string, string) (netproxy.Conn, error) { panic("injected underlay dial panic") }}
	s := &Server{logger: &logger, dialer: d, inFlightUnderlayKey: NewInFlightUnderlayKey(time.Second), udpEndpointPool: NewUdpEndpointPool()}
	key := &shadowsocks.Key{CipherConf: juicity.CipherConf, MasterKey: make([]byte, juicity.CipherConf.KeyLen)}
	salt := make([]byte, juicity.CipherConf.SaltLen)
	s.inFlightUnderlayKey.Store(inFlightKey(salt), &juicity.UnderlayAuth{Psk: key.MasterKey,
		Metadata: &trojanc.Metadata{Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: 9}, Network: "udp"}})
	packet, err := shadowsocks.EncryptUDPFromPool(key, []byte("payload"), salt, ciphers.JuicityReusedInfo)
	if err != nil {
		t.Fatal(err)
	}
	defer packet.Put()
	if err := s.dispatchNonQuicPacket(nil, packet, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9999}, newUnderlayWaitBudget()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-panicked:
	case <-time.After(time.Second):
		t.Fatal("underlay panic was not recovered and logged")
	}
}
