package server

import (
	"context"
	"sync/atomic"

	"github.com/daeuniverse/outbound/protocol/tuic/common"
	"github.com/daeuniverse/quic-go"
	"github.com/daeuniverse/quic-go/logging"
)

type congestionContextKey struct{}

type congestionConnection struct {
	conn      atomic.Pointer[quic.Connection]
	installed bool // Accessed only by the connection run loop.
}

func (s *Server) connContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, congestionContextKey{}, &congestionConnection{})
}

func (s *Server) congestionTracer(ctx context.Context, _ logging.Perspective, _ quic.ConnectionID) *logging.ConnectionTracer {
	pending, ok := ctx.Value(congestionContextKey{}).(*congestionConnection)
	if !ok || pending == nil {
		return nil
	}
	return &logging.ConnectionTracer{
		UpdatedMetrics: func(_ *logging.RTTStats, _, _ logging.ByteCount, _ int) {
			if pending.installed {
				return
			}
			if conn := pending.conn.Load(); conn != nil {
				// The fork invokes UpdatedMetrics synchronously on its run loop, unlike Accept.
				common.SetCongestionController(*conn, s.congestionControl, s.cwnd)
				pending.installed = true
			}
		},
	}
}
