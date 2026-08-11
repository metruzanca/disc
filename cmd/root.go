// Package cmd implements the disc CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "disc",
	Short: "A CLI for managing Discord servers via a bot.",
	Long: `Manage Discord servers, channels and roles.
Designed for automation and agent-based workflows.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(inviteCmd)
	rootCmd.AddCommand(channelCmd)
	rootCmd.AddCommand(eventCmd)
	rootCmd.AddCommand(roleCmd)
	rootCmd.AddCommand(configCmd)
}
