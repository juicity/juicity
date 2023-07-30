package log

import (
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
)

type Logger = zerolog.Logger

func NewLogger(timeFormat string) *Logger {
	// parse log level from config
	w := zerolog.NewConsoleWriter()
	w.TimeFormat = timeFormat
	l := zerolog.New(w)
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
