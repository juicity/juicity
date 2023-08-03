package main

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/juicity/juicity/common"
	"github.com/spf13/cobra"
)

var (
	genCertchainHashCmd = &cobra.Command{
		Use:                   "generate-certchain-hash [fullchain_file]",
		DisableFlagsInUseLine: true,
		Short:                 "To generate the hash of a full chain certificate.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(1)
			}
			hash, err := generateCertChainHash(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			fmt.Println(hash)
		},
	}
)

func generateCertChainHash(file string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	var rawCerts [][]byte
	for {
		block, rest := pem.Decode(b)
		if block == nil {
			break
		}
		rawCerts = append(rawCerts, block.Bytes)
		b = rest
	}
	if len(rawCerts) == 0 {
		return "", fmt.Errorf("not a certificate file")
	}
	return base64.StdEncoding.EncodeToString(common.GenerateCertChainHash(rawCerts)), nil
}

func init() {
	// cmds
	rootCmd.AddCommand(genCertchainHashCmd)
}
