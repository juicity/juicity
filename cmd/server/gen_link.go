package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/juicity/juicity/cmd/internal/shared"
	"github.com/spf13/cobra"
)

var (
	genLinkCmd = &cobra.Command{
		Use:                   "generate-sharinglink [config_file]",
		DisableFlagsInUseLine: true,
		Short:                 "To generate the sharing link from the config file.",
		Run: func(cmd *cobra.Command, args []string) {
			link, err := generateLink(shared.GetArguments())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(link)
		},
	}
)

func generateLink(arguments shared.Arguments) (string, error) {
	conf, err := arguments.GetConfig()
	if err != nil {
		return "", err
	}
	_, port, err := net.SplitHostPort(conf.Listen)
	if err != nil {
		return "", fmt.Errorf("parse 'listen': %w", err)
	}
	if len(conf.Users) == 0 {
		return "", fmt.Errorf("no users")
	}
	var (
		uuid     string
		password string
	)
	for uuid, password = range conf.Users {
		break
	}
	// Validate the cert and key.
	tlsCert, err := tls.LoadX509KeyPair(conf.Certificate, conf.PrivateKey)
	if err != nil {
		return "", err
	}
	cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return "", err
	}

	// Get cert hash.
	hash, err := generateCertChainHash(conf.Certificate)
	if err != nil {
		return "", fmt.Errorf("generateCertChainHash: %w", err)
	}
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	r := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Duration(timeout) * time.Millisecond,
			}
			return d.DialContext(ctx, "tcp", "208.67.222.222:53")
		},
	}

	// Get IP address.
	addrs, _ := r.LookupHost(ctx, "myip.opendns.com")
	if len(addrs) == 0 {
		http.DefaultClient.Timeout = timeout
		resp, err := http.Get("https://myipv4.p1.opendns.com/get_my_ip")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		var respBody struct {
			IP string `json:"ip"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return "", err
		}
		addrs = []string{respBody.IP}
	}
	link := url.URL{
		Scheme: "juicity",
		User:   url.UserPassword(uuid, password),
		Host:   net.JoinHostPort(addrs[0], port),
		RawQuery: url.Values{
			"congestion_control":      []string{"bbr"},
			"sni":                     []string{cert.Subject.CommonName},
			"pinned_certchain_sha256": []string{hash},
		}.Encode(),
	}
	return link.String(), nil
}

func init() {
	// cmds
	rootCmd.AddCommand(genLinkCmd)

	// flags
	shared.InitArgumentsFlags(genLinkCmd)
}
