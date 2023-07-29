package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/log"
	"github.com/juicity/juicity/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	logger  log.Logger
	cfgFile string

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juicity-server in the foreground.",
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
			lvl, err := zerolog.ParseLevel(conf.LogLevel)
			if err != nil {
				logger.Fatal().Err(err).Send()
			}

			*logger = logger.Level(lvl)

			go func() {
				if err := Serve(conf); err != nil {
					logger.Fatal().
						Err(err).
						Send()
				}
			}()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGILL)
			for sig := range sigs {
				logger.Warn().
					Str("signal", sig.String()).
					Msg("Exiting")
				return
			}
		},
	}
)

func Serve(conf *config.Config) error {
	fwmark, err := strconv.ParseUint(conf.Fwmark, 0, 32)
	if err != nil {
		return fmt.Errorf("parse fwmark: %w", err)
	}
	if uint64(fwmark) > math.MaxInt || uint64(fwmark) > math.MaxUint32 {
		return fmt.Errorf("fwmark is too large")
	}
	s, err := server.New(&server.Options{
		Users:             conf.Users,
		Certificate:       conf.Certificate,
		PrivateKey:        conf.PrivateKey,
		CongestionControl: conf.CongestionControl,
		Fwmark:            int(fwmark),
		SendThrough:       conf.SendThrough,
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
	// logger
	logger = log.AccessLogger()

	// cmds
	rootCmd.AddCommand(runCmd)

	// flags
	runCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file of juicity-server.")
}
