package log

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	logger := NewLogger(time.DateTime)

	const msg = "hello!"

	logger.Info().Msg(msg)
}
