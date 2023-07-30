package main

import (
	"github.com/juicity/juicity/internal/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "juicity-server [flags] [command [argument ...]]",
		Short:   "juicity-server is a quic-based proxy server.",
		Long:    "juicity-server is a quic-based proxy server.",
		Version: config.Version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
