package log

import (
	"path/filepath"
	"strconv"
	"sync"

	"github.com/rs/zerolog"
)

var (
	once   sync.Once
	logger *zerolog.Logger
)

func Logger() *zerolog.Logger {
	once.Do(func() {
		w := zerolog.NewConsoleWriter()
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
