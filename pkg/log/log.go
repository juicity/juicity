package log

import (
	"io"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger = zerolog.Logger

type Options struct {
	TimeFormat string
	LogFile    string
}

func NewLogger(opt *Options) *Logger {
	var writer io.Writer

	// parse log level from config
	c := zerolog.NewConsoleWriter()
	c.TimeFormat = opt.TimeFormat

	// https://github.com/natefinch/lumberjack
	if opt.LogFile != "" {
		f := &lumberjack.Logger{
			Filename:   opt.LogFile,
			MaxSize:    10,   // megabytes
			MaxBackups: 1,    // copies
			MaxAge:     1,    // days
			Compress:   true, // disabled by default
		}
		// set multiple write streams (default: [stdout, file])
		multi := zerolog.MultiLevelWriter(c, f)
		writer = multi
	} else {
		writer = &c
	}

	l := zerolog.New(writer)

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	l = l.With().Caller().Logger()
	if opt.TimeFormat != "" {
		l = l.With().Timestamp().Logger()
	}
	l.Level(zerolog.DebugLevel)
	return &l
}
