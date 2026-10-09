package server

import (
	"bufio"
	"context"
	"crypto/subtle"
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
	"github.com/daeuniverse/outbound/netproxy"
	"github.com/daeuniverse/outbound/pool"
	"github.com/daeuniverse/outbound/protocol/direct"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/tuic"
	"github.com/daeuniverse/outbound/protocol/tuic/common"
	"github.com/daeuniverse/quic-go"
	"github.com/google/uuid"
)

const (
	AuthenticateTimeout = 10 * time.Second
	AcceptTimeout       = AuthenticateTimeout
	inFlightUnderlayTtl = AuthenticateTimeout
)

var (
	authenticateTimeout     = AuthenticateTimeout
	ErrUnexpectedVersion    = fmt.Errorf("unexpected version")
	ErrUnexpectedCmdType    = fmt.Errorf("unexpected cmd type")
	ErrAuthenticationFailed = fmt.Errorf("authentication failed")
	ErrDisabledTrafficType  = fmt.Errorf("disabled traffic type")
)

type Options struct {
	Logger                *log.Logger
	Users                 map[string]string
	Certificate           string
	PrivateKey            string
	CongestionControl     string
	Fwmark                int
	SendThrough           string
	DialerLink            string
	DisableOutboundUdp443 bool
}

type Server struct {
	logger                 *log.Logger
	relay                  relay.Relay
	dialer                 netproxy.Dialer
	tlsConfig              *tls.Config
	maxOpenIncomingStreams int64
	congestionControl      string
	cwnd                   int
	users                  map[uuid.UUID]string
	fwmark                 int
	disableOutboundUdp443  bool
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
		dialer:                 d,
		tlsConfig:              &tls.Config{NextProtos: []string{"h3"}, MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert}},
		maxOpenIncomingStreams: 100,
		congestionControl:      opts.CongestionControl,
		cwnd:                   10,
		users:                  users,
		fwmark:                 opts.Fwmark,
		disableOutboundUdp443:  opts.DisableOutboundUdp443,
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
		Conn:        pktConn,
		ConnContext: s.connContext,
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
		Tracer:                         s.congestionTracer,
	})
	if err != nil {
		return err
	}
	go func() {
		waiting := newUnderlayWaitBudget()
		buf := pool.Get(consts.MaxUDPPayload + 1)
		defer buf.Put()
		for {
			n, addr, err := transport.ReadNonQUICPacket(context.Background(), buf)
			if err != nil {
				s.logger.Error().
					Err(err).
					Send()
				return
			}
			if n > consts.MaxUDPPayload {
				s.logger.Debug().Int("size", n).Msg("juicity dropped an oversized [underlay] packet")
				continue
			}
			if err := s.dispatchNonQuicPacket(&transport, buf[:n], addr.(*net.UDPAddr), waiting); err != nil {
				s.logger.Info().Err(err).Send()
			}
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

func (s *Server) handleConn(conn quic.Connection) (err error) {
	if pending, ok := conn.Context().Value(congestionContextKey{}).(*congestionConnection); ok && pending != nil {
		pending.conn.Store(&conn)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authCtx, authDone := context.WithTimeout(ctx, authenticateTimeout)
	defer authDone()
	authenticated := make(chan struct{})
	failAuth := context.AfterFunc(authCtx, func() {
		cancel()
		_ = conn.CloseWithError(tuic.AuthenticationFailed, "")
	})
	defer failAuth()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				s.logger.Error().Interface("panic", p).Msg("authentication stream panic")
				cancel()
				_ = conn.CloseWithError(tuic.AuthenticationFailed, "")
			}
		}()
		var (
			uniStream io.Reader
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
		if !failAuth() || authCtx.Err() != nil {
			cancel()
			_ = conn.CloseWithError(tuic.AuthenticationFailed, "")
			return
		}
		close(authenticated)
		authDone()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err = s.handleUnderlayAuth(ctx, uniStream); err != nil {
				if errors.Is(err, io.EOF) {
					s.logger.Debug().
						Err(err).
						Msg("handleUnderlayAuth: maybe is an old client?")
					return
				}
				s.logger.Error().
					Err(err).
					Msg("handleUnderlayAuth")
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
			defer func() {
				if p := recover(); p != nil {
					s.logger.Error().Interface("panic", p).Msg("proxy stream panic")
					stream.CancelRead(quic.StreamErrorCode(tuic.ProtocolError))
					stream.CancelWrite(quic.StreamErrorCode(tuic.ProtocolError))
				}
			}()
			if err := s.handleStream(ctx, authenticated, conn, stream); err != nil {
				s.logger.Warn().
					Err(err).
					Send()
			}
		}(stream)
	}
}

func (s *Server) handleStream(ctx context.Context, authenticated <-chan struct{}, conn quic.Connection, stream quic.Stream) error {
	lConn := juicity.NewConn(stream, nil, nil)
	defer lConn.Close()
	// Read the header and initiate the metadata
	_, err := lConn.Read(nil)
	if err != nil {
		return err
	}
	select {
	case <-authenticated:
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
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
		var n int
		var addr netip.AddrPort
		lConn := &juicity.PacketConn{Conn: lConn}
		buf := pool.Get(consts.MaxUDPPayload)
		defer func() {
			if buf != nil {
				pool.Put(buf)
			}
		}()
		for {
			_ = lConn.SetReadDeadline(time.Now().Add(consts.DefaultNatTimeout))
			n, addr, err = lConn.ReadFrom(buf)
			if err != nil {
				return fmt.Errorf("ReadFrom: %w", err)
			}
			if !s.disableOutboundUdp443 || addr.Port() != 443 {
				break
			}
			s.logger.Debug().Str("target", addr.String()).Str("source", source).
				Msg("juicity blocked a [udp] frame")
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
		defer c.Close()
		rConn := c.(netproxy.PacketConn)
		_ = rConn.SetWriteDeadline(time.Now().Add(consts.DefaultNatTimeout)) // should keep consistent
		_, err = rConn.WriteTo(buf[:n], addr.String())
		pool.Put(buf)
		buf = nil
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
			s.disableOutboundUdp443,
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

func (s *Server) handleConnAuth(authCtx context.Context, conn quic.Connection) (uuid *uuid.UUID, reader io.Reader, err error) {
	uniStream, err := conn.AcceptUniStream(authCtx)
	if err != nil {
		return nil, nil, err
	}
	if deadline, ok := authCtx.Deadline(); ok {
		if err := uniStream.SetReadDeadline(deadline); err != nil {
			return nil, nil, err
		}
		defer uniStream.SetReadDeadline(time.Time{})
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
				if subtle.ConstantTimeCompare(token[:], authenticate.TOKEN[:]) == 1 {
					return &authenticate.UUID, r, nil
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
func (s *Server) handleUnderlayAuth(ctx context.Context, uniStream io.Reader) (err error) {
	// Read an auth from the connection.
	var auth juicity.UnderlayAuth
	if _, err = auth.Unpack(uniStream); err != nil {
		return err
	}

	// Store the key.
	s.inFlightUnderlayKey.Store(inFlightKey(auth.IV), &auth)
	return nil
}
