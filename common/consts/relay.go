package consts

import (
	"time"
)

const (
	EthernetMtu       = 1500
	DefaultNatTimeout = 3 * time.Minute
	DnsQueryTimeout   = 17 * time.Second // RFC 5452
)
