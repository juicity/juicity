package server

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/juicity/juicity/common/consts"
	"github.com/juicity/juicity/internal/relay"
	"github.com/juicity/juicity/pkg/log"

	"github.com/daeuniverse/outbound/dialer"
	"github.com/daeuniverse/softwind/ciphers"
	"github.com/daeuniverse/softwind/netproxy"
	"github.com/daeuniverse/softwind/pkg/fastrand"
	"github.com/daeuniverse/softwind/pool"
	"github.com/daeuniverse/softwind/protocol/direct"
	"github.com/daeuniverse/softwind/protocol/juicity"
	"github.com/daeuniverse/softwind/protocol/shadowsocks"
	"github.com/daeuniverse/softwind/protocol/tuic"
	"github.com/daeuniverse/softwind/protocol/tuic/common"
	"github.com/google/uuid"
	"github.com/mzz2017/quic-go"
)

const (
	AuthenticateTimeout = 10 * time.Second
	AcceptTimeout       = AuthenticateTimeout
	inFlightUnderlayTtl = AuthenticateTimeout
)

var (
	ErrUnexpectedVersion    = fmt.Errorf("unexpected version")
	ErrUnexpectedCmdType    = fmt.Errorf("unexpected cmd type")
	ErrAuthenticationFailed = fmt.Errorf("authentication failed")
)

type Options struct {
	Logger            *log.Logger
	Users             map[string]string
	Certificate       string
	PrivateKey        string
	CongestionControl string
	Fwmark            int
	SendThrough       string
	DialerLink        string
}

type Server struct {
	logger                 *log.Logger
	relay                  relay.Relay
	dialer                 netproxy.ContextDialer
	tlsConfig              *tls.Config
	maxOpenIncomingStreams int64
	congestionControl      string
	cwnd                   int
	users                  map[uuid.UUID]string
	fwmark                 int
	inFlightUnderlayKey    *InFlightUnderlayKey
	udpEndpointPool        *UdpEndpointPool
}

func New(opts *Options) (*Server, error) {
	users := map[uuid.UUID]string{}
	for _uuid, password := range opts.Users {
		id, err := uuid.Parse(_uuid)
		if err != nil {
			return nil, fmt.Errorf("parse uuid(%v): %w", _uuid, err)
		}
		users[id] = password
	}
	cert, err := tls.LoadX509KeyPair(opts.Certificate, opts.PrivateKey)
	if err != nil {
		return nil, err
	}
	var d netproxy.Dialer
	uesFullconeDialer := opts.DialerLink == ""
	switch {
	case opts.SendThrough != "":
		lAddr, err := netip.ParseAddr(opts.SendThrough)
		if err != nil {
			return nil, fmt.Errorf("parse send_through: %w", err)
		}
		d = direct.NewDirectDialerLaddr(uesFullconeDialer, lAddr)
	case uesFullconeDialer:
		d = direct.FullconeDirect
	default:
		d = direct.SymmetricDirect
	}
	if opts.DialerLink != "" {
		var property *dialer.Property
		if d, property, err = dialer.NewNetproxyDialerFromLink(d, &dialer.ExtraOption{
			AllowInsecure:     false,
			TlsImplementation: "",
			UtlsImitate:       "",
		}, opts.DialerLink); err != nil {
			return nil, fmt.Errorf("parse DialerLink: %w", err)
		}
		opts.Logger.Info().
			Str("name", property.Name).
			Str("proto", property.Protocol).
			Str("addr", property.Address).
			Msg("Dial use given dialer")
	}

	return &Server{
		logger:                 opts.Logger,
		relay:                  relay.NewRelay(opts.Logger),
		dialer:                 &netproxy.ContextDialerConverter{Dialer: d},
		tlsConfig:              &tls.Config{NextProtos: []string{"h3"}, MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert}},
		maxOpenIncomingStreams: 100,
		congestionControl:      opts.CongestionControl,
		cwnd:                   10,
		users:                  users,
		fwmark:                 opts.Fwmark,
		inFlightUnderlayKey:    NewInFlightUnderlayKey(inFlightUnderlayTtl),
		udpEndpointPool:        NewUdpEndpointPool(),
	}, nil
}

func (s *Server) Serve(addr string) (err error) {
	quicMaxOpenIncomingStreams := int64(s.maxOpenIncomingStreams)

	pktConn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	transport := quic.Transport{
		Conn: pktConn,
	}
	listener, err := transport.Listen(s.tlsConfig, &quic.Config{
		InitialStreamReceiveWindow:     common.InitialStreamReceiveWindow,
		MaxStreamReceiveWindow:         common.MaxStreamReceiveWindow,
		InitialConnectionReceiveWindow: common.InitialConnectionReceiveWindow,
		MaxConnectionReceiveWindow:     common.MaxConnectionReceiveWindow,
		MaxIncomingStreams:             quicMaxOpenIncomingStreams,
		MaxIncomingUniStreams:          quicMaxOpenIncomingStreams,
		KeepAlivePeriod:                10 * time.Second,
		DisablePathMTUDiscovery:        false,
		EnableDatagrams:                false,
		CapabilityCallback:             nil,
	})
	if err != nil {
		return err
	}
	go func() {
		buf := pool.GetFullCap(consts.EthernetMtu)
		defer buf.Put()
		for {
			n, addr, err := transport.ReadNonQUICPacket(context.Background(), buf)
			if err != nil {
				s.logger.Error().
					Err(err).
					Send()
				return
			}
			newBuf := pool.Get(n)
			copy(newBuf, buf)
			go func(transport *quic.Transport, buf pool.PB, ulAddr *net.UDPAddr) {
				defer buf.Put()
				if err := s.handleNonQuicPacket(transport, buf, ulAddr); err != nil {
					s.logger.Info().
						Err(err).
						Send()
				}
			}(&transport, newBuf, addr.(*net.UDPAddr))
		}
	}()
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}
		go func(conn quic.Connection) {
			if err := s.handleConn(conn); err != nil {
				var netError net.Error
				if errors.As(err, &netError) && netError.Timeout() {
					return // ignore i/o timeout
				}
				s.logger.Warn().
					Err(err).
					Send()
			}
		}(conn)
	}
}

func (s *Server) handleNonQuicPacket(transport *quic.Transport, buf []byte, ulAddr *net.UDPAddr) (err error) {
	if len(buf) < juicity.CipherConf.SaltLen {
		return fmt.Errorf("insuffient [underlay] data: len %v", len(buf))
	}
	lAddr := ulAddr.AddrPort()
	// source ip/port -> dst mapping.
	endpoint, isNew, err := s.udpEndpointPool.GetOrCreate(lAddr, &UdpEndpointOptions{
		Handler: func(data []byte, from netip.AddrPort, metadata any) error {
			masterKey := metadata.([]byte)
			salt := pool.Get(juicity.CipherConf.SaltLen)
			defer salt.Put()
			_, _ = fastrand.Read(salt)
			salt[0] = 0
			salt[1] = 0
			buf, err := shadowsocks.EncryptUDPFromPool(&shadowsocks.Key{
				CipherConf: juicity.CipherConf,
				MasterKey:  masterKey,
			}, data, salt, ciphers.JuicityReusedInfo)
			if err != nil {
				return err
			}
			defer buf.Put()
			_, err = transport.WriteTo(buf, ulAddr)
			return err
		},
		NatTimeout: 0,
		GetDialOption: func() (*DialOption, error) {
			iv := buf[:juicity.CipherConf.SaltLen]
			auth := s.inFlightUnderlayKey.Evict(inFlightKey(iv))
			if auth == nil {
				return nil, fmt.Errorf("[underlay] auth fail")
			}
			return &DialOption{
				Target:   net.JoinHostPort(auth.Metadata.Hostname, strconv.Itoa(int(auth.Metadata.Port))),
				Dialer:   s.dialer,
				Metadata: auth.Psk,
			}, nil
		},
	})
	if err != nil {
		return err
	}
	if !isNew {
		masterKey := endpoint.Metadata.([]byte)
		decrypted, err := shadowsocks.DecryptUDPFromPool(&shadowsocks.Key{
			CipherConf: juicity.CipherConf,
			MasterKey:  masterKey,
		}, buf, ciphers.JuicityReusedInfo)
		if err != nil {
			return err
		}
		defer decrypted.Put()
		buf = decrypted
	} else {
		s.logger.Debug().
			Str("target", endpoint.DialTarget).
			Str("source", lAddr.String()).
			Msg("juicity performed an [underlay] request")
	}
	if _, err = endpoint.WriteTo(buf, endpoint.DialTarget); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleConn(conn quic.Connection) (err error) {
	common.SetCongestionController(conn, s.congestionControl, s.cwnd)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authCtx, authDone := context.WithTimeout(ctx, AuthenticateTimeout)
	defer authDone()
	go func() {
		var (
			uniStream quic.ReceiveStream
			err       error
		)
		if _, uniStream, err = s.handleConnAuth(authCtx, conn); err != nil {
			s.logger.Warn().
				Err(err).
				Msg("handleAuth")
			cancel()
			_ = conn.CloseWithError(tuic.AuthenticationFailed, "")
			return
		}
		authDone()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err = s.handleUnderlayAuth(ctx, uniStream); err != nil {
				s.logger.Warn().
					Err(err).
					Msg("handleUnderlayAuth")
				cancel()
				return
			}
		}
	}()
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return err
		}
		go func(stream quic.Stream) {
			if err = s.handleStream(ctx, authCtx, conn, stream); err != nil {
				s.logger.Warn().
					Err(err).
					Send()
			}
		}(stream)
	}
}

func (s *Server) handleStream(ctx context.Context, authCtx context.Context, conn quic.Connection, stream quic.Stream) error {
	lConn := juicity.NewConn(stream, nil, nil)
	defer lConn.Close()
	// Read the header and initiate the metadata
	_, err := lConn.Read(nil)
	if err != nil {
		return err
	}
	<-authCtx.Done()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	mdata := lConn.Metadata
	source := conn.RemoteAddr().String()
	switch mdata.Network {
	case "tcp":
		target := net.JoinHostPort(mdata.Hostname, strconv.Itoa(int(mdata.Port)))
		s.logger.Debug().
			Str("target", target).
			Str("source", source).
			Msg("juicity received a [tcp] request")
		magicNetwork := netproxy.MagicNetwork{
			Network: "tcp",
			Mark:    uint32(s.fwmark),
		}
		ctx, cancel := context.WithTimeout(context.Background(), consts.DefaultDialTimeout)
		defer cancel()
		rConn, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), target)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				s.logger.Debug().
					Err(err).
					Send()
				return nil // ignore i/o timeout
			}
			return err
		}
		defer rConn.Close()
		if err = s.relay.RelayTCP(lConn, rConn); err != nil {
			var netErr net.Error
			if errors.Is(err, io.EOF) || (errors.As(err, &netErr) && netErr.Timeout()) || strings.HasSuffix(err.Error(), "with error code 0") {
				return nil // ignore i/o timeout
			}
			return fmt.Errorf("relay tcp error: %w", err)
		}
	case "udp":
		// can dial any target
		lConn := &juicity.PacketConn{Conn: lConn}
		buf := pool.GetFullCap(consts.EthernetMtu)
		defer pool.Put(buf)
		_ = lConn.SetReadDeadline(time.Now().Add(consts.DefaultNatTimeout))
		n, addr, err := lConn.ReadFrom(buf)
		if err != nil {
			return fmt.Errorf("ReadFrom: %w", err)
		}

		magicNetwork := netproxy.MagicNetwork{
			Network: "udp",
			Mark:    uint32(s.fwmark),
		}
		ctx, cancel := context.WithTimeout(context.Background(), consts.DefaultDialTimeout)
		defer cancel()
		c, err := s.dialer.DialContext(ctx, magicNetwork.Encode(), addr.String())
		s.logger.Debug().
			Str("target", addr.String()).
			Str("source", source).
			Msg("juicity received a [udp] request")
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				return nil // ignore i/o timeout
			}
			return fmt.Errorf("Dial: %w", err)
		}
		rConn := c.(netproxy.PacketConn)
		_ = rConn.SetWriteDeadline(time.Now().Add(consts.DefaultNatTimeout)) // should keep consistent
		_, err = rConn.WriteTo(buf[:n], addr.String())
		if err != nil {
			if errors.Is(err, net.ErrWriteToConnected) {
				s.logger.Warn().
					Err(err).
					Msg("relayConnToUDP")
			}
			return fmt.Errorf("WriteTo: %w", err)
		}
		if err = s.relay.RelayUoT(
			rConn,
			lConn,
			len(buf),
		); err != nil {
			var netErr net.Error
			if errors.Is(err, io.EOF) || (errors.As(err, &netErr) && netErr.Timeout()) || strings.HasSuffix(err.Error(), "with error code 0") {
				return nil // ignore i/o timeout
			}
			return fmt.Errorf("relay udp error: %w", err)
		}
	default:
		return fmt.Errorf("unexpected network: %v", mdata.Network)
	}
	return nil
}

func (s *Server) handleConnAuth(authCtx context.Context, conn quic.Connection) (uuid *uuid.UUID, uniStream quic.ReceiveStream, err error) {
	uniStream, err = conn.AcceptUniStream(authCtx)
	if err != nil {
		return nil, nil, err
	}
	r := bufio.NewReader(uniStream)
	v, err := r.Peek(1)
	if err != nil {
		return nil, nil, err
	}
	switch v[0] {
	case juicity.Version0:
		commandHead, err := tuic.ReadCommandHead(r)
		if err != nil {
			return nil, nil, fmt.Errorf("ReadCommandHead: %w", err)
		}
		switch commandHead.TYPE {
		case tuic.AuthenticateType:
			authenticate, err := tuic.ReadAuthenticateWithHead(commandHead, r)
			if err != nil {
				return nil, nil, fmt.Errorf("ReadAuthenticateWithHead: %w", err)
			}
			var token [32]byte
			if password, ok := s.users[authenticate.UUID]; ok {
				token, err = tuic.GenToken(conn.ConnectionState(), authenticate.UUID, password)
				if err != nil {
					return nil, nil, fmt.Errorf("GenToken: %w", err)
				}
				if token == authenticate.TOKEN {
					return &authenticate.UUID, uniStream, nil
				} else {
					_ = conn.CloseWithError(tuic.AuthenticationFailed, "")
				}
			}
			return nil, nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, authenticate.UUID)
		default:
			return nil, nil, fmt.Errorf("%w: %v", ErrUnexpectedCmdType, commandHead.TYPE)
		}
	default:
		return nil, nil, fmt.Errorf("%w: %v", ErrUnexpectedVersion, v)
	}
}
func (s *Server) handleUnderlayAuth(ctx context.Context, uniStream quic.ReceiveStream) (err error) {
	// Read an auth from the connection.
	var auth juicity.UnderlayAuth
	if _, err = auth.Unpack(uniStream); err != nil {
		return err
	}

	// Store the key.
	s.inFlightUnderlayKey.Store(inFlightKey(auth.IV), &auth)
	return nil
}
