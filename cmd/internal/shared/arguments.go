package shared

import (
	"fmt"
	"sync"
	"time"

	"github.com/juicity/juicity/common/consts"
	"github.com/juicity/juicity/config"
	"github.com/juicity/juicity/pkg/log"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type Arguments struct {
	CfgFile             string
	disableTimestamp    bool
	LogDisableTimestamp bool
	LogOutput           string
	LogDisableColor     bool
	LogFile             string
	LogFileFormat       string
	LogMaxSize          int
	LogMaxBackups       int
	LogMaxAge           int
	LogCompress         bool
}

var (
	defaultArguments Arguments
	onceArguments    sync.Once
)

func GetArguments() Arguments {
	onceArguments.Do(func() {
		a := &defaultArguments
		if a.disableTimestamp {
			a.LogDisableTimestamp = true
		}
	})
	return defaultArguments
}

func (a *Arguments) GetLogger(logLevel string) (*log.Logger, error) {
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("ParseLevel: %w", err)
	}
	logTimeFormat := time.DateTime
	if a.LogDisableTimestamp {
		logTimeFormat = ""
	}
	logger := log.NewLogger(&log.Options{
		Output:     a.LogOutput,
		TimeFormat: logTimeFormat,
		FileFormat: a.LogFileFormat,
		NoColor:    a.LogDisableColor,
		File:       a.LogFile,
		MaxSize:    a.LogMaxSize,
		MaxBackups: a.LogMaxBackups,
		MaxAge:     a.LogMaxAge,
		Compress:   a.LogCompress,
	})
	*logger = logger.Level(lvl)
	return logger, nil
}

func (a *Arguments) GetConfig() (*config.Config, error) {
	if a.CfgFile == "" {
		return nil, fmt.Errorf("argument \"--config\" or \"-c\" is required but not provided")
	}

	// Read config from --config cfgFile.
	conf, err := config.ReadConfig(a.CfgFile)
	if err != nil {
		return nil, fmt.Errorf("ReadConfig: %w", err)
	}
	return conf, nil
}

func InitArgumentsFlags(cmd *cobra.Command) {
	// flags
	cmd.PersistentFlags().StringVarP(&defaultArguments.CfgFile, "config", "c", "", "specify config file path")
	// log-related flags
	cmd.PersistentFlags().StringVarP(&defaultArguments.LogOutput, "log-output", "", "console", "specify the log outputs; options: [console|file|console,file]")
	cmd.PersistentFlags().BoolVarP(&defaultArguments.LogDisableColor, "log-disable-color", "", false, "disable colorful log output")
	// Deprecated: Use log-disable-timestamp instead.
	cmd.PersistentFlags().BoolVarP(&defaultArguments.disableTimestamp, "disable-timestamp", "", false, "deprecated; use log-disable-timestamp instead")
	cmd.PersistentFlags().BoolVarP(&defaultArguments.LogDisableTimestamp, "log-disable-timestamp", "", false, "disable timestamp")
	cmd.PersistentFlags().StringVarP(&defaultArguments.LogFile, "log-file", "", "/var/log/juicity-client.log", "log file path to write")
	cmd.PersistentFlags().StringVarP(&defaultArguments.LogFileFormat, "log-file-format", "", "raw", "specify log format; options: [raw|json]")
	cmd.PersistentFlags().IntVarP(&defaultArguments.LogMaxSize, "log-file-max-size", "", consts.LogMaxSize, "specify maximum size of the log file before it gets rotated; unit: MB")
	cmd.PersistentFlags().IntVarP(&defaultArguments.LogMaxBackups, "log-file-max-backups", "", consts.LogMaxBackups, "specify the maximum number of old log files to retain")
	cmd.PersistentFlags().IntVarP(&defaultArguments.LogMaxAge, "log-file-max-age", "", consts.LogMaxAge, "specify the maximum number of days to retain old log files based on the timestamp encoded in their filename; unit: day")
	cmd.PersistentFlags().BoolVarP(&defaultArguments.LogCompress, "log-file-compress", "", consts.LogCompress, "enable log compression; default: true")
}
