package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/juicity/juicity/cmd/internal/shared"
	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/log"
	"github.com/juicity/juicity/server"

	"github.com/spf13/cobra"
)

var (
	logger = log.NewLogger(&log.Options{
		TimeFormat: time.DateTime,
	})

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print out version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("juicity-client version %v\n", config.Version)
			fmt.Printf("go version %v %v/%v\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
			fmt.Printf("CGO_ENABLED: %v\n", cgoEnabled)
		},
	}

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juicity-server in the foreground.",
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

			go func() {
				if err := Serve(conf); err != nil {
					logger.Fatal().
						Err(err).
						Send()
				}
			}()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGILL)
			for sig := range sigs {
				logger.Warn().
					Str("signal", sig.String()).
					Msg("Exiting")
				return
			}
		},
	}
)

func Serve(conf *config.Config) (err error) {
	var fwmark uint64
	if conf.Fwmark != "" {
		fwmark, err = strconv.ParseUint(conf.Fwmark, 0, 32)
		if err != nil {
			return fmt.Errorf("parse fwmark: %w", err)
		}
		if fwmark > math.MaxInt || fwmark > math.MaxUint32 {
			return fmt.Errorf("fwmark is too large")
		}
	}
	s, err := server.New(&server.Options{
		Logger:                logger,
		Users:                 conf.Users,
		Certificate:           conf.Certificate,
		PrivateKey:            conf.PrivateKey,
		CongestionControl:     conf.CongestionControl,
		Fwmark:                int(fwmark),
		SendThrough:           conf.SendThrough,
		DialerLink:            conf.DialerLink,
		DisableOutboundUdp443: conf.DisableOutboundUdp443,
	})
	if err != nil {
		return err
	}
	if conf.Listen == "" {
		return fmt.Errorf(`"Listen" is required`)
	}
	logger.Info().Msg("Listen at " + conf.Listen)
	if err = s.Serve(conf.Listen); err != nil {
		return err
	}
	return nil
}

func init() {
	// cmds
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)

	// flags
	shared.InitArgumentsFlags(runCmd)
}
