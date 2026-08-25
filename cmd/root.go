package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var agentMode bool

var rootCmd = &cobra.Command{
	Use:   "disc",
	Short: "A CLI for managing Discord servers via a bot.",
	Long: `Manage Discord servers, channels, roles, and scheduled events.
Designed for automation and agent-based workflows.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&agentMode, "agent", false, "Disable all interactive prompts; error on missing required params")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(inviteCmd)
	rootCmd.AddCommand(channelCmd)
	rootCmd.AddCommand(eventCmd)
	rootCmd.AddCommand(roleCmd)
}
