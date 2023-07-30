package log

import (
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	once   sync.Once
	logger *zerolog.Logger
)

type Logger = *zerolog.Logger

func AccessLogger() *zerolog.Logger {
	once.Do(func() {
		w := zerolog.NewConsoleWriter()
		w.TimeFormat = time.DateTime
		l := zerolog.New(w)
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			return filepath.Base(file) + ":" + strconv.Itoa(line)
		}
		l = l.With().Caller().Logger().With().Timestamp().Logger()
		l.Level(zerolog.DebugLevel)
		logger = &l
	})
	return logger
}
