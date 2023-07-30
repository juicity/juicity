package log

import (
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger = zerolog.Logger

func NewLogger(timeFormat string) *Logger {
	// parse log level from config
	c := zerolog.NewConsoleWriter()
	c.TimeFormat = timeFormat

	// https://github.com/natefinch/lumberjack
	f := &lumberjack.Logger{
		Filename:   "app.log",
		MaxSize:    10,   // megabytes
		MaxBackups: 1,    // copies
		MaxAge:     1,    // days
		Compress:   true, // disabled by default
	}

	// set multiple write streams (default: [stdout, file])
	multi := zerolog.MultiLevelWriter(c, f)
	l := zerolog.New(multi)

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	l = l.With().Caller().Logger()
	if timeFormat != "" {
		l = l.With().Timestamp().Logger()
	}
	l.Level(zerolog.DebugLevel)
	return &l
}
