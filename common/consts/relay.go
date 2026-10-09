package consts

import (
	"time"
)

// default value for relay (internal/relay)
const (
	EthernetMtu       = 1500
	MaxUDPPayload     = 1<<16 - 1 // Juicity frames use a uint16 payload length.
	DefaultNatTimeout = 3 * time.Minute
	DnsQueryTimeout   = 17 * time.Second // RFC 5452
)
