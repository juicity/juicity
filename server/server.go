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

	"github.com/daeuniverse/softwind/netproxy"
	"github.com/daeuniverse/softwind/pool"
	"github.com/daeuniverse/softwind/protocol/direct"
	"github.com/daeuniverse/softwind/protocol/juicity"
	"github.com/daeuniverse/softwind/protocol/tuic"
	"github.com/daeuniverse/softwind/protocol/tuic/common"
	"github.com/google/uuid"
	"github.com/mzz2017/quic-go"
)

const (
	AuthenticateTimeout = 10 * time.Second
	AcceptTimeout       = AuthenticateTimeout
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
	dialer := direct.FullconeDirect
	if opts.SendThrough != "" {
		lAddr, err := netip.ParseAddr(opts.SendThrough)
		if err != nil {
			return nil, fmt.Errorf("parse send_through: %w", err)
		}
		dialer = direct.NewDirectDialerLaddr(true, lAddr)
	}
	return &Server{
		logger: opts.Logger,
		relay:  relay.NewRelay(opts.Logger),
		dialer: dialer,
		tlsConfig: &tls.Config{
			NextProtos:   []string{"h3"}, // h3 only.
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{cert},
		},
		maxOpenIncomingStreams: 100,
		congestionControl:      opts.CongestionControl,
		cwnd:                   10,
		users:                  users,
		fwmark:                 opts.Fwmark,
	}, nil
}

func (s *Server) Serve(addr string) (err error) {
	quicMaxOpenIncomingStreams := int64(s.maxOpenIncomingStreams)

	listener, err := quic.ListenAddr(addr, s.tlsConfig, &quic.Config{
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
	common.SetCongestionController(conn, s.congestionControl, s.cwnd)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authCtx, authDone := context.WithTimeout(ctx, AuthenticateTimeout)
	defer authDone()
	go func() {
		if _, err := s.handleAuth(authCtx, conn); err != nil {
			s.logger.Warn().
				Err(err).
				Msg("handleAuth")
			cancel()
			_ = conn.CloseWithError(tuic.AuthenticationFailed, "")
			return
		}
		authDone()
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
	defer stream.Close()
	lConn := juicity.NewConn(stream, nil, nil)
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
			Msg("juicity received a tcp request")
		magicNetwork := netproxy.MagicNetwork{
			Network: "tcp",
			Mark:    uint32(s.fwmark),
		}
		rConn, err := s.dialer.Dial(magicNetwork.Encode(), target)
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
		c, err := s.dialer.Dial(magicNetwork.Encode(), addr.String())
		s.logger.Debug().
			Str("target", addr.String()).
			Str("source", source).
			Msg("juicity received a udp request")
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

func (s *Server) handleAuth(ctx context.Context, conn quic.Connection) (uuid *uuid.UUID, err error) {
	uniStream, err := conn.AcceptUniStream(ctx)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(uniStream)
	v, err := r.Peek(1)
	if err != nil {
		return nil, err
	}
	switch v[0] {
	case juicity.Version0:
		commandHead, err := tuic.ReadCommandHead(r)
		if err != nil {
			return nil, fmt.Errorf("ReadCommandHead: %w", err)
		}
		switch commandHead.TYPE {
		case tuic.AuthenticateType:
			authenticate, err := tuic.ReadAuthenticateWithHead(commandHead, r)
			if err != nil {
				return nil, fmt.Errorf("ReadAuthenticateWithHead: %w", err)
			}
			var token [32]byte
			if password, ok := s.users[authenticate.UUID]; ok {
				token, err = tuic.GenToken(conn.ConnectionState(), authenticate.UUID, password)
				if err != nil {
					return nil, fmt.Errorf("GenToken: %w", err)
				}
				if token == authenticate.TOKEN {
					return &authenticate.UUID, nil
				} else {
					_ = conn.CloseWithError(tuic.AuthenticationFailed, "")
				}
			}
			return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, authenticate.UUID)
		default:
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedCmdType, commandHead.TYPE)
		}
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnexpectedVersion, v)
	}
}
