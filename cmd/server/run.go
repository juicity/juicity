package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mzz2017/juice/cmd/internal"
	"github.com/mzz2017/juice/config"
	"github.com/mzz2017/juice/pkg/log"
	"github.com/mzz2017/juice/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	runCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file of juice-server.")
	runCmd.PersistentFlags().BoolVarP(&disableTimestamp, "disable-timestamp", "", false, "Disable timestamp.")
}

var (
	cfgFile          string
	disableTimestamp bool

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juice-server in the foreground.",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgFile == "" {
				log.Logger().
					Fatal().
					Msg("Argument \"--config\" or \"-c\" is required but not provided.")
			}

			// Require "sudo" if necessary.
			internal.AutoSu()

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
			log.Logger().Level(lvl)

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
	s, err := server.New(&server.Options{
		Users:       conf.Users,
		Certificate: conf.Certificate,
		PrivateKey:  conf.PrivateKey,
	})
	if err != nil {
		return err
	}
	if err = s.Serve(conf.Listen); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}
