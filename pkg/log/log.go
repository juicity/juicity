package log

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger = zerolog.Logger

type Options struct {
	TimeFormat    string
	LogFile       string
	NoColor       bool
	JsonLogFormat bool
}

func NewLogger(opt *Options) *Logger {
	var writer io.Writer

	// default writer: ConsoleWriter
	// additional writer options: [jsonWriter, fileWriter]

	// consoleWriter
	c := &zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: opt.TimeFormat,
		NoColor:    opt.NoColor,
	}

	// jsonWriter
	if opt.JsonLogFormat {
		writer = zerolog.MultiLevelWriter(os.Stdout)
	} else {
		writer = zerolog.MultiLevelWriter(c)
	}

	// fileWriter (additional log stream)
	// https://github.com/natefinch/lumberjack
	if opt.LogFile != "" {
		// this will write log to file in stdout format
		f := zerolog.ConsoleWriter{
			Out: &lumberjack.Logger{
				Filename:   opt.LogFile,
				MaxSize:    10,   // megabytes
				MaxBackups: 1,    // copies
				MaxAge:     1,    // days
				Compress:   true, // disabled by default
			},
			TimeFormat: opt.TimeFormat,
			NoColor:    opt.NoColor,
		}

		// set multiple write streams (default: [stdout, file])
		writer = zerolog.MultiLevelWriter(c, f)
	}

	// fileWriter + jsonWriter
	if opt.LogFile != "" && opt.JsonLogFormat {
		f := &lumberjack.Logger{
			Filename:   opt.LogFile,
			MaxSize:    10,   // megabytes
			MaxBackups: 1,    // copies
			MaxAge:     1,    // days
			Compress:   true, // disabled by default
		}

		// set multiple write streams (default: [stdout, file])
		writer = zerolog.MultiLevelWriter(os.Stdout, f)
	}

	logger := zerolog.New(writer)

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	logger = logger.With().Caller().Logger()
	if opt.TimeFormat != "" {
		logger = logger.With().Timestamp().Logger()
	}
	logger.Level(zerolog.DebugLevel)

	return &logger
}
