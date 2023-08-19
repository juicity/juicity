package main

import (
	"os"

	_ "github.com/daeuniverse/outbound/dialer/http"
	_ "github.com/daeuniverse/outbound/dialer/juicity"
	_ "github.com/daeuniverse/outbound/dialer/shadowsocks"
	_ "github.com/daeuniverse/outbound/dialer/shadowsocksr"
	_ "github.com/daeuniverse/outbound/dialer/socks"
	_ "github.com/daeuniverse/outbound/dialer/trojan"
	_ "github.com/daeuniverse/outbound/dialer/tuic"
	_ "github.com/daeuniverse/outbound/dialer/v2ray"
	_ "github.com/daeuniverse/outbound/transport/simpleobfs"
	_ "github.com/daeuniverse/outbound/transport/tls"
	_ "github.com/daeuniverse/outbound/transport/ws"
	_ "github.com/daeuniverse/softwind/protocol/juicity"
	_ "github.com/daeuniverse/softwind/protocol/shadowsocks"
	_ "github.com/daeuniverse/softwind/protocol/trojanc"
	_ "github.com/daeuniverse/softwind/protocol/tuic"
	_ "github.com/daeuniverse/softwind/protocol/vless"
	_ "github.com/daeuniverse/softwind/protocol/vmess"
)

func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
