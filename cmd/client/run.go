package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/daeuniverse/softwind/protocol"
	"github.com/daeuniverse/softwind/protocol/juicity"
	gliderLog "github.com/nadoo/glider/pkg/log"
	"github.com/spf13/cobra"

	"github.com/juicity/juicity/cmd/internal/shared"
	"github.com/juicity/juicity/common"
	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/client/dialer"
	"github.com/juicity/juicity/pkg/log"
	"github.com/juicity/juicity/server"
)

var (
	logger = log.NewLogger(&log.Options{
		TimeFormat: time.DateTime,
	})

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juicity-client in the foreground.",
		Run: func(cmd *cobra.Command, args []string) {
			arguments := shared.GetArguments()
			// Config.
			conf, err := arguments.GetConfig()
			if err != nil {
				logger.Fatal().
					Err(err).
					Msg("Failed to read config")
			}

			// Logger.
			if logger, err = arguments.GetLogger(conf.LogLevel); err != nil {
				logger.Fatal().
					Err(err).
					Msg("Failed to init logger")
			}
			gliderLog.SetLogger(logger)

			// QUIC_GO_ENABLE_GSO
			gso, _ := strconv.ParseBool(os.Getenv("QUIC_GO_ENABLE_GSO"))
			logger.Info().
				Bool("Requested QUIC_GO_ENABLE_GSO", gso).
				Send()

			go func() {
				if err := Serve(conf); err != nil {
					logger.Fatal().Err(err).Send()
				}
			}()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGILL)
			for sig := range sigs {
				logger.Warn().Str("signal", sig.String()).Msg("Exiting")
				return
			}
		},
	}
)

func Serve(conf *config.Config) error {
	if conf.Sni == "" {
		conf.Sni, _, _ = net.SplitHostPort(conf.Server)
	}
	tlsConfig := &tls.Config{
		NextProtos:         []string{"h3"},
		MinVersion:         tls.VersionTLS13,
		ServerName:         conf.Sni,
		InsecureSkipVerify: conf.AllowInsecure,
	}
	if conf.PinnedCertChainSha256 != "" {
		pinnedHash, err := base64.URLEncoding.DecodeString(conf.PinnedCertChainSha256)
		if err != nil {
			pinnedHash, err = base64.StdEncoding.DecodeString(conf.PinnedCertChainSha256)
			if err != nil {
				pinnedHash, err = hex.DecodeString(conf.PinnedCertChainSha256)
				if err != nil {
					return fmt.Errorf("failed to decode PinnedCertChainSha256")
				}
			}
		}
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if !bytes.Equal(common.GenerateCertChainHash(rawCerts), pinnedHash) {
				return fmt.Errorf("pinned hash of cert chain does not match")
			}
			return nil
		}
	}
	d, err := juicity.NewDialer(dialer.NewClientDialer(conf), protocol.Header{
		ProxyAddress: conf.Server,
		Feature1:     conf.CongestionControl,
		TlsConfig:    tlsConfig,
		User:         conf.Uuid,
		Password:     conf.Password,
		IsClient:     true,
		Flags:        0,
	})
	if err != nil {
		return err
	}
	s, err := server.NewMixed("mixed://"+conf.Listen, d)
	if err != nil {
		return err
	}
	if conf.Listen == "" {
		return fmt.Errorf(`"Listen" is required`)
	}
	logger.Info().Msg("Listen http and socks5 at " + conf.Listen)
	s.ListenAndServe()
	return nil
}

func init() {
	// cmds
	rootCmd.AddCommand(runCmd)
	shared.InitArgumentsFlags(runCmd)
}
