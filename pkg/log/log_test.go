package log

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	logger := NewLogger(&Options{
		TimeFormat: time.DateTime,
		LogFile:    "../../app.log",
	})

	const msg = "hello!"

	logger.Info().Msg(msg)
}
