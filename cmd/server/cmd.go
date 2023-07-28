package main

import (
	"github.com/mzz2017/juicity/config"
	"github.com/spf13/cobra"
)

var (
	Version = "unknown"
	rootCmd = &cobra.Command{
		Use:     "juicity-server [flags] [command [argument ...]]",
		Short:   "juicity-server is a quic-based proxy server.",
		Long:    "juicity-server is a quic-based proxy server.",
		Version: Version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

func init() {
	config.Version = Version
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
