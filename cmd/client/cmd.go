package main

import (
	"github.com/spf13/cobra"
)

var (
	cgoEnabled = 0
	rootCmd    = &cobra.Command{
		Use:   "juicity-client [flags] [command [argument ...]]",
		Short: "juicity-client is a quic-based proxy client.",
		Long:  "juicity-client is a quic-based proxy client.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
)

func init() {
	rootCmd.AddCommand()
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
