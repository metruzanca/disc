package cmd

import (
	"github.com/metruzanca/disc/internal/discord"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all bots and servers under disc management",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		bf, err := loadBots()
		if err != nil {
			return err
		}
		if bf.Bots == nil || len(bf.Bots) == 0 {
			util.Yellow.Println("Status: Not configured")
			util.Bold.Println("Run disc server new to add a bot and server.")
			return nil
		}

		util.Green.Println("Status: Connected")
		util.Bold.Printf("Bots (%d):\n", len(bf.Bots))
		for botID, bot := range bf.Bots {
			activeMarker := ""
			if bf.ActiveBot == botID {
				activeMarker = " [active bot]"
			}
			client, err := discord.New(bot.Token)
			if err != nil {
				util.Red.Printf("  Bot %s%s — token invalid (rotated?)\n", botID, activeMarker)
				continue
			}
			if err := client.WaitReady(); err != nil {
				client.Close()
				util.Red.Printf("  Bot %s%s — token invalid (rotated?)\n", botID, activeMarker)
				continue
			}

			user := client.User()
			guilds := client.Guilds()
			client.Close()

			util.Green.Printf("  Bot: %s%s\n", user.Username, activeMarker)
			util.Cyan.Printf("    ID: %s\n", botID)

			if len(guilds) == 0 {
				util.Dim.Println("    Servers: (bot is not in any servers)")
				continue
			}

			var managed, unmanaged []string
			for _, g := range guilds {
				_, managedByUs := bot.Servers[g.ID]
				if managedByUs {
					managed = append(managed, g.Name)
				} else {
					unmanaged = append(unmanaged, g.Name)
				}
			}

			if len(managed) > 0 {
				util.Green.Printf("    Managed servers (%d):\n", len(managed))
				for _, name := range managed {
					activeSrv := ""
					if bf.ActiveBot == botID && bf.ActiveServer != "" {
						for sid := range bot.Servers {
							if bot.Servers[sid].Name == name && sid == bf.ActiveServer {
								activeSrv = " [active]"
								break
							}
						}
					}
					util.Cyan.Printf("      - %s%s\n", name, activeSrv)
				}
			}
			if len(unmanaged) > 0 {
				util.Dim.Printf("    Other servers (%d):\n", len(unmanaged))
				for _, name := range unmanaged {
					util.Dim.Printf("      - %s\n", name)
				}
			}
		}
		return nil
	},
}
