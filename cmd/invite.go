package cmd

import (
	"fmt"

	"github.com/metruzanca/disc/internal/discord"
	"github.com/spf13/cobra"
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Get the invite link for the active bot",
	Long:  "Generate an OAuth2 invite link to add the active bot to a Discord server. The link includes permissions for managing channels and roles.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := resolveActiveClient()
		if err != nil {
			return err
		}
		defer client.Close()

		fmt.Println("Invite your bot to a server using this link:")
		fmt.Println()
		fmt.Println(discord.InviteLink(client.User().ID))
		return nil
	},
}
