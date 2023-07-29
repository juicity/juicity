package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	logger := NewLogger()

	const msg = "hello!"

	logger.Info().Msg(msg)
}
