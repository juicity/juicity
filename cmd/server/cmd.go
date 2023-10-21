package main

import (
	"github.com/spf13/cobra"
)

var (
	cgoEnabled = 0
	rootCmd    = &cobra.Command{
		Use:   "juicity-server [flags] [command [argument ...]]",
		Short: "juicity-server is a quic-based proxy server.",
		Long:  "juicity-server is a quic-based proxy server.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
