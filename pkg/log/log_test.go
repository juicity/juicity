package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	logger := AccessLogger()

	const msg = "hello!"

	logger.Info().Msg(msg)
}
