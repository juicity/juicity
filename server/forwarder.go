package server

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"

	"github.com/juicity/juicity/common/consts"
	"github.com/juicity/juicity/internal/relay"
	"github.com/juicity/juicity/pkg/log"

	"github.com/daeuniverse/softwind/netproxy"
	"github.com/daeuniverse/softwind/pool"
	concPool "github.com/sourcegraph/conc/pool"
)

type ForwarderOptions struct {
	Logger     *log.Logger
	Dialer     netproxy.Dialer
	LocalAddr  string
	RemoteAddr string
}

type Forwarder struct {
	ctx    context.Context
	cancel func()

	ForwarderOptions
	relay relay.Relay

	relayTcp    bool
	relayUdp    bool
	tcpListener *net.TCPListener
	udpListener *net.UDPConn

	udpEndpointPool *UdpEndpointPool
}

func NewForwarder(opts ForwarderOptions) (*Forwarder, error) {
	ctx, cancel := context.WithCancel(context.Background())
	fields := strings.Split(opts.LocalAddr, "/")
	var (
		isTcp bool
		isUdp bool
	)
	if len(fields) > 1 {
		opts.LocalAddr = fields[0]
		for i := 1; i < len(fields); i++ {
			switch fields[i] {
			case "tcp":
				isTcp = true
			case "udp":
				isUdp = true
			}
		}
	} else {
		isTcp = true
		isUdp = true
	}
	return &Forwarder{
		ctx:              ctx,
		cancel:           cancel,
		ForwarderOptions: opts,
		relay:            relay.NewRelay(opts.Logger),
		relayTcp:         isTcp,
		relayUdp:         isUdp,
		tcpListener:      nil,
		udpListener:      nil,
		udpEndpointPool:  NewUdpEndpointPool(),
	}, nil
}

func (s *Forwarder) Serve() (err error) {
	defer func() {
		if err != nil {
			s.Close()
		}
	}()
	var network string
	if s.relayTcp && s.relayUdp {
		network = "tcp/udp"
	} else if s.relayTcp {
		network = "tcp"
	} else {
		network = "udp"
	}

	s.Logger.Info().Msgf("Forward local %v <-%v-> remote %v", s.LocalAddr, network, s.RemoteAddr)

	wg := concPool.New().WithErrors().WithContext(s.ctx).WithCancelOnError()
	if s.relayTcp {
		tcpListener, err := net.Listen("tcp", s.LocalAddr)
		if err != nil {
			return err
		}
		s.tcpListener = tcpListener.(*net.TCPListener)

		wg.Go(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					s.Close()
				}
			}()
			for {
				conn, err := s.tcpListener.Accept()
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				go func(lConn net.Conn) {
					defer lConn.Close()
					rConn, err := s.Dialer.Dial("tcp", s.RemoteAddr)
					if err != nil {
						s.Logger.Info().
							Err(err).
							Str("target", s.RemoteAddr).
							Msg("Failed to dial TCP")
						return
					}
					s.Logger.Info().Msgf("Forward %v <-tcp-> %v", lConn.RemoteAddr().String(), s.RemoteAddr)
					if err := s.relay.RelayTCP(lConn, rConn); err != nil {
						var netError net.Error
						if errors.As(err, &netError) && netError.Timeout() {
							return // ignore i/o timeout
						}
						s.Logger.Warn().
							Err(err).
							Send()
					}
				}(conn)
			}
		})
	}
	if s.relayUdp {
		uAddr, err := net.ResolveUDPAddr("udp", s.LocalAddr)
		if err != nil {
			return err
		}
		udpListener, err := net.ListenUDP("udp", uAddr)
		if err != nil {
			return err
		}
		s.udpListener = udpListener
		wg.Go(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					s.Close()
				}
			}()
			buf := pool.GetFullCap(consts.EthernetMtu)
			defer buf.Put()
			for {
				n, addr, err := s.udpListener.ReadFromUDPAddrPort(buf)
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				newBuf := pool.Get(n)
				copy(newBuf, buf[:n])
				go func(buf pool.PB, lAddr netip.AddrPort) {
					defer buf.Put()
					endpoint, isNew, err := s.udpEndpointPool.GetOrCreate(lAddr, &UdpEndpointOptions{
						Handler: func(data []byte, from netip.AddrPort, metadata any) error {
							_, err := s.udpListener.WriteToUDPAddrPort(data, lAddr)
							return err
						},
						NatTimeout: consts.DefaultNatTimeout,
						GetDialOption: func() (*DialOption, error) {
							return &DialOption{
								Target: s.RemoteAddr,
								Dialer: s.Dialer,
							}, nil
						},
					})
					if err != nil {
						s.Logger.Info().
							Err(err).
							Str("source", lAddr.String()).
							Str("target", s.RemoteAddr).
							Msg("Failed to dial UDP")
						return
					}
					if isNew {
						s.Logger.Info().Msgf("Forward %v <-udp-> %v", addr.String(), s.RemoteAddr)
					}
					if _, err = endpoint.WriteTo(buf, s.RemoteAddr); err != nil {
						s.Logger.Info().
							Err(err).
							Str("source", lAddr.String()).
							Str("target", s.RemoteAddr).
							Msg("Failed to write UDP data")
						return
					}
				}(newBuf, addr)
			}
		})
	}
	if err = wg.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *Forwarder) Close() error {
	select {
	case <-s.ctx.Done():
	default:
		s.cancel()
		if l := s.udpListener; l != nil {
			l.Close()
		}
		if l := s.tcpListener; l != nil {
			l.Close()
		}
	}
	return nil
}
