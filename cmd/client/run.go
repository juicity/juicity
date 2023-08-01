package main

import (
	"crypto/tls"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mzz2017/softwind/protocol"
	"github.com/mzz2017/softwind/protocol/juicity"
	gliderLog "github.com/nadoo/glider/pkg/log"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/client/dialer"
	"github.com/juicity/juicity/pkg/log"
	"github.com/juicity/juicity/server"
)

var (
	logger           *log.Logger
	cfgFile          string
	disableTimestamp bool
	logFile          string
	noLogColor       bool
	logFormat        string

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juicity-client in the foreground.",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgFile == "" {
				logger.Fatal().
					Msg("Argument \"--config\" or \"-c\" is required but not provided.")
			}

			// Read config from --config cfgFile.
			conf, err := config.ReadConfig(cfgFile)
			if err != nil {
				logger.Fatal().
					Err(err).
					Msg("Failed to read config")
			}

			// Logger.
			lvl, err := zerolog.ParseLevel(conf.LogLevel)
			if err != nil {
				logger.Fatal().Err(err).Send()
			}
			if lvl <= zerolog.InfoLevel {
				flag := 0
				if !disableTimestamp {
					flag = stdlog.Ltime | stdlog.Ldate
				}
				gliderLog.Set(true, flag)
			}
			timeFormat := time.DateTime
			if disableTimestamp {
				timeFormat = ""
			}
			logger = log.NewLogger(&log.Options{
				TimeFormat: timeFormat,
				LogFile:    logFile,
				NoColor:    noLogColor,
				LogFormat:  logFormat,
			})
			*logger = logger.Level(lvl)

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
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGILL)
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
	d, err := juicity.NewDialer(dialer.NewClientDialer(conf), protocol.Header{
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
	logger.Info().Msg("Listen http and socks5 at " + conf.Listen)
	s.ListenAndServe()
	return nil
}

func init() {
	// cmds
	rootCmd.AddCommand(runCmd)

	// flags
	runCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "specify config file path")
	runCmd.PersistentFlags().BoolVarP(&disableTimestamp, "disable-timestamp", "", false, "disable timestamp")
	runCmd.PersistentFlags().StringVarP(&logFile, "log-file", "", "", "write logs to file")
	runCmd.PersistentFlags().StringVarP(&logFormat, "log-format", "", "raw", "specify log format; options: [raw,json]; default: raw")
	runCmd.PersistentFlags().BoolVarP(&noLogColor, "log-disable-color", "", false, "disable colorful log output")
}
