package main

import (
	"fmt"
	"os"

	"github.com/juicity/juicity/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "juicity-server [flags] [command [argument ...]]",
		Short: "juicity-server is a quic-based proxy server.",
		Long:  "juicity-server is a quic-based proxy server.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print out version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("juicity-server version %v\n%v\n", config.Version, config.Runtime)
			if val, isSet := os.LookupEnv("CGO_ENALBED"); !isSet {
				fmt.Print("CGO_ENALBED: NOT SET\n")
			} else {
				fmt.Printf("CGO_ENALBED: %v\n", val)
			}
		},
	})
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
