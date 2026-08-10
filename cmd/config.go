package cmd

import (
	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage disc configuration",
	Long:  "View and modify disc configuration settings.",
}

var configSetServerCmd = &cobra.Command{
	Use:   "set-server <id>",
	Short: "Set the default server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.ServerID = args[0]
		if err := config.Save(cfg); err != nil {
			return err
		}

		name := args[0]
		if token := cfg.Token; token != "" {
			client, err := discord.New(token)
			if err == nil {
				if err := client.WaitReady(); err == nil {
					if g, err := client.Session().Guild(args[0]); err == nil {
						name = g.Name
					}
				}
				client.Close()
			}
		}

		util.Green.Printf("Done! Default server set to %s (%s)\n", name, args[0])
		return nil
	},
}

var configClearServerCmd = &cobra.Command{
	Use:   "clear-server",
	Short: "Clear the default server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.ServerID = ""
		if err := config.Save(cfg); err != nil {
			return err
		}
		util.Green.Println("Default server cleared.")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetServerCmd)
	configCmd.AddCommand(configClearServerCmd)
}
