package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/log"
	"github.com/juicity/juicity/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	runCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file of juicity-server.")
}

var (
	cfgFile string

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juicity-server in the foreground.",
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
	if uint64(conf.Fwmark) > math.MaxInt || uint64(conf.Fwmark) > math.MaxUint32 {
		return fmt.Errorf("fwmark is too large")
	}
	s, err := server.New(&server.Options{
		Users:             conf.Users,
		Certificate:       conf.Certificate,
		PrivateKey:        conf.PrivateKey,
		CongestionControl: conf.CongestionControl,
		Fwmark:            conf.Fwmark,
	})
	if err != nil {
		return err
	}
	if conf.Listen == "" {
		return fmt.Errorf(`"Listen" is required`)
	}
	log.Logger().Info().
		Msg("Listen at " + conf.Listen)
	if err = s.Serve(conf.Listen); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}
