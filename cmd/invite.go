package cmd

import (
	"fmt"

	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
	"github.com/spf13/cobra"
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Get the invite link to add the bot to a server",
	Long:  "Generate an OAuth2 invite link to add the bot to a Discord server. The link includes permissions for managing channels and roles.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		token := cfg.Token
		if token == "" {
			return fmt.Errorf("no bot token configured; run disc init")
		}

		client, err := discord.New(token)
		if err != nil {
			return err
		}
		defer client.Close()
		if err := client.WaitReady(); err != nil {
			return err
		}

		fmt.Println("Invite your bot to a server using this link:")
		fmt.Println()
		fmt.Println(discord.InviteLink(client.User().ID))
		return nil
	},
}
