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
	TimeFormat       string
	Format           string
	NoColor          bool
	EnableFileWriter bool
	File             string
	MaxSize          int
	MaxBackups       int
	MaxAge           int
	Compress         bool
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
	if opt.Format == "json" {
		writer = zerolog.MultiLevelWriter(os.Stdout)
	} else {
		writer = zerolog.MultiLevelWriter(c)
	}

	// fileWriter (additional log stream)
	// https://github.com/natefinch/lumberjack
	if opt.EnableFileWriter && opt.File != "" {
		// this will write log to file in stdout format
		f := zerolog.ConsoleWriter{
			Out: &lumberjack.Logger{
				Filename:   opt.File,       // path
				MaxSize:    opt.MaxSize,    // megabytes
				MaxBackups: opt.MaxBackups, // copies
				MaxAge:     opt.MaxAge,     // days
				Compress:   opt.Compress,   // enable by default
			},
			TimeFormat: opt.TimeFormat,
			NoColor:    opt.NoColor,
		}

		// set multiple write streams (default: [stdout, file])
		writer = zerolog.MultiLevelWriter(c, f)
	}

	// fileWriter + jsonWriter
	if opt.EnableFileWriter && opt.File != "" && opt.Format == "json" {
		f := &lumberjack.Logger{
			Filename:   opt.File,       // path
			MaxSize:    opt.MaxSize,    // megabytes
			MaxBackups: opt.MaxBackups, // copies
			MaxAge:     opt.MaxAge,     // days
			Compress:   opt.Compress,   // enable by default
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
