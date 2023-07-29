package main

import (
	"github.com/juicity/juicity/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "juicity-client [flags] [command [argument ...]]",
		Short:   "juicity-client is a quic-based proxy client.",
		Long:    "juicity-client is a quic-based proxy client.",
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
