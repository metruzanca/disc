package cmd

import (
	"fmt"

	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show CLI status and available servers",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		token := cfg.Token
		if token == "" {
			util.Yellow.Println("Status: Not configured")
			util.Bold.Println("Run disc init to set up your bot token.")
			return nil
		}

		client, err := discord.New(token)
		if err != nil {
			return err
		}
		defer client.Close()
		if err := client.WaitReady(); err != nil {
			return err
		}

		user := client.User()
		util.Green.Println("Status: Connected")
		util.Bold.Printf("Bot: %s\n", user.Username)

		guilds := client.Guilds()
		if len(guilds) == 0 {
			util.Yellow.Println("Servers (0):")
			return nil
		}

		util.Bold.Printf("\nServers (%d):\n", len(guilds))
		for _, g := range guilds {
			util.Cyan.Printf("  - %s ", g.Name)
			util.Dim.Printf("(ID: %s)", g.ID)
			if cfg.ServerID != "" && g.ID == cfg.ServerID {
				util.Green.Printf(" [default]")
			}
			fmt.Println()
		}
		return nil
	},
}
