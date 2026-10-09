package server

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"sync"

	"github.com/daeuniverse/outbound/ciphers"
	"github.com/daeuniverse/outbound/pkg/fastrand"
	"github.com/daeuniverse/outbound/pool"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/shadowsocks"
	"github.com/daeuniverse/quic-go"
)

const (
	// Only unknown credentials wait: leave live endpoints and authenticated first packets outside this budget.
	maxPendingUnderlayHandlers = 256
	// A single source IP can queue a small burst without occupying the global waiting budget.
	maxPendingUnderlayPerIP = 8
)

type underlayWaitBudget struct {
	mu      sync.Mutex
	pending chan struct{}
	byIP    map[netip.Addr]int
}

func newUnderlayWaitBudget() *underlayWaitBudget {
	return &underlayWaitBudget{pending: make(chan struct{}, maxPendingUnderlayHandlers), byIP: make(map[netip.Addr]int)}
}

func (b *underlayWaitBudget) acquire(ip netip.Addr) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.byIP[ip] >= maxPendingUnderlayPerIP {
		return false
	}
	select {
	case b.pending <- struct{}{}:
		b.byIP[ip]++
		return true
	default:
		return false
	}
}

func (b *underlayWaitBudget) release(ip netip.Addr) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.byIP[ip]--
	if b.byIP[ip] == 0 {
		delete(b.byIP, ip)
	}
	<-b.pending
}

func (s *Server) recoverUnderlayPanic() {
	if p := recover(); p != nil {
		s.logger.Error().Interface("panic", p).Msg("underlay handler panic: packet dropped")
	}
}

func (s *Server) dispatchNonQuicPacket(transport *quic.Transport, buf []byte, ulAddr *net.UDPAddr, waiting *underlayWaitBudget) error {
	defer s.recoverUnderlayPanic()
	if len(buf) < juicity.CipherConf.SaltLen {
		return fmt.Errorf("insufficient [underlay] data: len %v", len(buf))
	}
	lAddr := ulAddr.AddrPort()
	if endpoint, ok := s.udpEndpointPool.Get(lAddr); ok {
		return forwardUnderlayPacket(endpoint, buf)
	}
	k := inFlightKey(buf[:juicity.CipherConf.SaltLen])
	if auth := s.inFlightUnderlayKey.Lookup(k); auth != nil {
		payload, err := s.verifyUnderlayPacket(k, auth, buf, lAddr)
		if err == ErrDisabledTrafficType {
			return nil
		}
		if err != nil {
			return err
		}
		go s.handleUnderlayPayload(transport, payload, ulAddr, auth)
		return nil
	}
	ip := lAddr.Addr().Unmap()
	if !waiting.acquire(ip) {
		s.logger.Debug().Str("source", ip.String()).Msg("juicity dropped an [underlay] packet: credential wait budget full")
		return nil
	}
	packet := pool.Get(len(buf))
	copy(packet, buf)
	go func() {
		defer s.recoverUnderlayPanic()
		defer packet.Put()
		auth := func() *juicity.UnderlayAuth {
			defer waiting.release(ip)
			return s.inFlightUnderlayKey.Wait(k)
		}()
		if auth == nil {
			return
		}
		payload, err := s.verifyUnderlayPacket(k, auth, packet, lAddr)
		if err == ErrDisabledTrafficType {
			return
		}
		if err != nil {
			s.logger.Info().Err(err).Send()
			return
		}
		s.handleUnderlayPayload(transport, payload, ulAddr, auth)
	}()
	return nil
}

func (s *Server) verifyUnderlayPacket(k inFlightKey, auth *juicity.UnderlayAuth, buf []byte, source netip.AddrPort) (pool.PB, error) {
	if s.disableOutboundUdp443 && auth.Metadata.Port == 443 {
		s.logger.Debug().
			Str("target", net.JoinHostPort(auth.Metadata.Hostname, strconv.Itoa(int(auth.Metadata.Port)))).
			Str("source", source.String()).Msg("juicity blocked an [underlay] request")
		return nil, ErrDisabledTrafficType
	}
	payload, err := shadowsocks.DecryptUDPFromPool(&shadowsocks.Key{
		CipherConf: juicity.CipherConf, MasterKey: auth.Psk,
	}, buf, ciphers.JuicityReusedInfo)
	if err != nil {
		return nil, err
	}
	if !s.inFlightUnderlayKey.Consume(k, auth) {
		payload.Put()
		return nil, ErrAuthenticationFailed
	}
	return payload, nil
}

func forwardUnderlayPacket(endpoint *UdpEndpoint, buf []byte) error {
	payload, err := shadowsocks.DecryptUDPFromPool(&shadowsocks.Key{
		CipherConf: juicity.CipherConf, MasterKey: endpoint.Metadata.([]byte),
	}, buf, ciphers.JuicityReusedInfo)
	if err != nil {
		return err
	}
	defer payload.Put()
	_, err = endpoint.WriteTo(payload, endpoint.DialTarget)
	return err
}

func (s *Server) handleUnderlayPayload(transport *quic.Transport, payload pool.PB, ulAddr *net.UDPAddr, auth *juicity.UnderlayAuth) {
	defer s.recoverUnderlayPanic()
	defer payload.Put()
	lAddr := ulAddr.AddrPort()
	endpoint, isNew, err := s.udpEndpointPool.GetOrCreate(lAddr, &UdpEndpointOptions{
		Logger: s.logger,
		Handler: func(data []byte, from netip.AddrPort, metadata any) error {
			salt := pool.Get(juicity.CipherConf.SaltLen)
			defer salt.Put()
			_, _ = fastrand.Read(salt)
			salt[0], salt[1] = 0, 0
			buf, err := shadowsocks.EncryptUDPFromPool(&shadowsocks.Key{
				CipherConf: juicity.CipherConf, MasterKey: metadata.([]byte),
			}, data, salt, ciphers.JuicityReusedInfo)
			if err != nil {
				return err
			}
			defer buf.Put()
			_, err = transport.WriteTo(buf, ulAddr)
			return err
		},
		GetDialOption: func() (*DialOption, error) {
			return &DialOption{
				Target: net.JoinHostPort(auth.Metadata.Hostname, strconv.Itoa(int(auth.Metadata.Port))),
				Dialer: s.dialer, Metadata: auth.Psk,
			}, nil
		},
	})
	if err != nil {
		s.logger.Info().Err(err).Send()
		return
	}
	if !bytes.Equal(endpoint.Metadata.([]byte), auth.Psk) {
		s.logger.Info().Err(ErrAuthenticationFailed).Send()
		return
	}
	if isNew {
		s.logger.Debug().Str("target", endpoint.DialTarget).Str("source", lAddr.String()).
			Msg("juicity performed an [underlay] request")
	}
	if _, err := endpoint.WriteTo(payload, endpoint.DialTarget); err != nil {
		s.logger.Info().Err(err).Send()
	}
}
