package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/juicity/juicity/common/consts"
	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/log"
	"github.com/juicity/juicity/server"
)

var (
	logger           *log.Logger
	cfgFile          string
	disableTimestamp bool
	logDisableColor  bool
	logFile          string
	logFormat        string
	logMaxSize       int
	logMaxBackups    int
	logMaxAge        int
	logCompress      bool

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "To run juicity-server in the foreground.",
		Run: func(cmd *cobra.Command, args []string) {
			logger = log.NewLogger(&log.Options{
				TimeFormat:       time.DateTime,
				EnableFileWriter: false,
			})

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

			// Logger
			lvl, err := zerolog.ParseLevel(conf.LogLevel)
			if err != nil {
				logger.Fatal().Err(err).Send()
			}
			timeFormat := time.DateTime
			if disableTimestamp {
				timeFormat = ""
			}
			logger = log.NewLogger(&log.Options{
				TimeFormat: timeFormat,
				NoColor:    logDisableColor,
				File:       logFile,
				Format:     logFormat,
				MaxSize:    logMaxSize,
				MaxBackups: logMaxBackups,
				MaxAge:     logMaxAge,
				Compress:   logCompress,
			})
			*logger = logger.Level(lvl)

			// QUIC_GO_ENABLE_GSO
			gso, _ := strconv.ParseBool(os.Getenv("QUIC_GO_ENABLE_GSO"))
			logger.Info().
				Bool("Requested QUIC_GO_ENABLE_GSO", gso).
				Send()

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
		Logger:            logger,
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
	// cmds
	rootCmd.AddCommand(runCmd)

	// flags
	runCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "specify config file path")
	runCmd.PersistentFlags().BoolVarP(&disableTimestamp, "disable-timestamp", "", false, "disable timestamp; default: false")
	// log-related flags
	runCmd.PersistentFlags().StringVarP(&logFile, "log-file", "", "", "write logs to file; default: /var/log/juicity/juicity.log")
	runCmd.PersistentFlags().StringVarP(&logFormat, "log-format", "", "raw", "specify log format; options: [raw,json]; default: raw")
	runCmd.PersistentFlags().BoolVarP(&logDisableColor, "log-disable-color", "", false, "disable colorful log output")
	runCmd.PersistentFlags().IntVarP(&logMaxSize, "log-max-size", "", consts.LogMaxSize, "specify maximum size in megabytes of the log file before it gets rotated; default: 10 megabytes")
	runCmd.PersistentFlags().IntVarP(&logMaxBackups, "log-max-backups", "", consts.LogMaxBackups, "specify the maximum number of old log files to retain; default: 1")
	runCmd.PersistentFlags().IntVarP(&logMaxAge, "log-max-age", "", consts.LogMaxAge, "specify the maximum number of days to retain old log files based on the timestamp encoded in their filename; default: 1 day")
	runCmd.PersistentFlags().BoolVarP(&logCompress, "log-compress", "", consts.LogCompress, "enable log compression; default: true")
}
