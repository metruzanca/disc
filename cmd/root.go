// Package cmd implements the disc CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// agentMode disables all interactive prompts for the invocation. Agents pass
// --agent to guarantee the CLI never blocks on stdin; missing required params
// then produce a descriptive error instead.
var agentMode bool

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
	rootCmd.PersistentFlags().BoolVar(&agentMode, "agent", false, "Disable all interactive prompts; error on missing required params")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(inviteCmd)
	rootCmd.AddCommand(channelCmd)
	rootCmd.AddCommand(eventCmd)
	rootCmd.AddCommand(roleCmd)
	rootCmd.AddCommand(configCmd)
}
