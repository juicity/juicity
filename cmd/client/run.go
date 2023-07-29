package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	stdlog "log"

	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/log"
	"github.com/juicity/juicity/server"
	"github.com/mzz2017/softwind/protocol"
	"github.com/mzz2017/softwind/protocol/direct"
	"github.com/mzz2017/softwind/protocol/juicity"
	gliderLog "github.com/nadoo/glider/pkg/log"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	runCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file of juicity-server.")
}

var (
	cfgFile          string
	disableTimestamp bool

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juicity-client in the foreground.",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgFile == "" {
				log.Logger().
					Fatal().
					Msg("Argument \"--config\" or \"-c\" is required but not provided.")
			}

			// Read config from --config cfgFile.
			conf, err := config.ReadConfig(cfgFile)
			if err != nil {
				log.Logger().
					Fatal().
					Err(err).
					Msg("Failed to read config")
			}
			lvl, err := zerolog.ParseLevel(conf.LogLevel)
			if err != nil {
				log.Logger().
					Fatal().
					Err(err).
					Send()
			}
			if lvl <= zerolog.InfoLevel {
				gliderLog.Set(true, stdlog.Ltime)
			}
			*log.Logger() = log.Logger().Level(lvl)

			go func() {
				if err := Serve(conf); err != nil {
					log.Logger().Fatal().
						Err(err).
						Send()
				}
			}()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGILL)
			for sig := range sigs {
				log.Logger().Warn().
					Str("signal", sig.String()).
					Msg("Exiting")
				return
			}
		},
	}
)

func Serve(conf *config.Config) error {
	if conf.Sni == "" {
		conf.Sni, _, _ = net.SplitHostPort(conf.Server)
	}
	d, err := juicity.NewDialer(direct.SymmetricDirect, protocol.Header{
		ProxyAddress: conf.Server,
		Feature1:     conf.CongestionControl,
		TlsConfig: &tls.Config{
			NextProtos:         []string{"h3"},
			MinVersion:         tls.VersionTLS13,
			ServerName:         conf.Sni,
			InsecureSkipVerify: conf.AllowInsecure,
		},
		User:     conf.Uuid,
		Password: conf.Password,
		IsClient: true,
		Flags:    0,
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
	log.Logger().Info().
		Msg("Listen http and socks5 at " + conf.Listen)
	if err = s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}
