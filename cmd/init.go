package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/huh"
	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var (
	initFlags = struct {
		token          string
		configLocation string
		serverID       string
	}{}
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up disc with your bot token",
	Long: `Initialize disc by providing your Discord bot token.
After validating the token, this command records your preferred config
location (local directory or disc config directory) and prints an
invite link to add the bot to a server.

In agent mode (--agent), --config-location is required (local or global)
and --server-id is required when using global mode and the bot is in
multiple servers.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readToken()
		if err != nil {
			return err
		}
		if strings.TrimSpace(token) == "" {
			return fmt.Errorf("no token provided")
		}

		loc := initFlags.configLocation
		if loc == "" {
			if agentMode {
				return fmt.Errorf("--config-location is required in agent mode (local or global)")
			}
			loc, err = promptConfigLocation()
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					util.Yellow.Println("Aborted.")
					return nil
				}
				return err
			}
		}
		if loc != "local" && loc != "global" {
			return fmt.Errorf("--config-location must be 'local' or 'global'")
		}

		client, err := discord.New(strings.TrimSpace(token))
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}
		defer client.Close()

		if err := client.WaitReady(); err != nil {
			return err
		}

		guilds := client.Guilds()

		var serverID string
		if initFlags.serverID != "" {
			serverID = initFlags.serverID
		} else if len(guilds) == 1 {
			serverID = guilds[0].ID
		} else if len(guilds) > 1 {
			if agentMode {
				return fmt.Errorf("--server-id is required when the bot is in multiple servers (pass --server-id, set DISCORD_SERVER_ID, or run non-agent)")
			}
			serverID, err = promptServerSelection(guilds)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					util.Yellow.Println("Aborted.")
					return nil
				}
				return err
			}
		}

		if loc == "global" && serverID == "" {
			util.Yellow.Println("Invite your bot to a server using this link:")
			fmt.Println()
			fmt.Println(discord.InviteLink(client.User().ID))
			fmt.Println()
			return fmt.Errorf("global config mode requires the bot to be in a server; add the bot and re-run disc init")
		}

		if err := writeTokenAndPrefs(token, serverID, loc); err != nil {
			return err
		}

		util.Green.Printf("Token saved.\n")
		if loc == "local" {
			util.Green.Printf("Config location: local directory (%s)\n", config.EnvFileName)
		} else {
			sd, _ := config.ServerDir(serverID)
			util.Green.Printf("Config location: disc config directory (%s)\n", sd)
		}
		fmt.Println()
		fmt.Println("Invite your bot to a server using this link:")
		fmt.Println()
		fmt.Println(discord.InviteLink(client.User().ID))
		fmt.Println()
		if serverID != "" {
			fmt.Println("Your server config will be saved at:")
			if loc == "global" {
				p, _ := config.ServerConfigFile(serverID)
				fmt.Printf("  %s\n", p)
				fmt.Println("(Edit with 'disc config role/channel add', then apply with 'disc config push')")
			} else {
				fmt.Println("  ./server.json")
				fmt.Println("(Commit to version control, or run 'disc config pull' to snapshot the live server)")
			}
		}
		return nil
	},
}

func promptConfigLocation() (string, error) {
	var loc string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Where should disc save your server config?").
			Description("Local: commit server.json to version control (best for managing servers as code).\nGlobal: save to ~/.config/disc/servers/<id>/ — all configs in one place (best for simple usage).").
			Options(
				huh.NewOption("Local directory", "local"),
				huh.NewOption("disc config directory", "global"),
			).
			Value(&loc),
	))
	if err := form.Run(); err != nil {
		return "", err
	}
	return loc, nil
}

func promptServerSelection(guilds []*discordgo.Guild) (string, error) {
	var serverID string
	options := make([]huh.Option[string], len(guilds))
	for i, g := range guilds {
		options[i] = huh.NewOption(g.Name, g.ID)
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Which server should this config be for?").
			Description("The bot is in multiple servers. Pick the one you want to manage with disc.").
			Options(options...).
			Value(&serverID),
	))
	if err := form.Run(); err != nil {
		return "", err
	}
	return serverID, nil
}

func writeTokenAndPrefs(token, serverID, loc string) error {
	prefs := &config.Prefs{ServerConfigLocation: loc}
	if err := config.SavePrefs(prefs); err != nil {
		return fmt.Errorf("failed to save prefs: %w", err)
	}

	if loc == "global" {
		if err := config.SetServerEnvVar(serverID, config.TokenEnv, token); err != nil {
			return fmt.Errorf("failed to write token to global env: %w", err)
		}
		if err := config.SetServerEnvVar(serverID, config.ServerIDEnv, serverID); err != nil {
			return fmt.Errorf("failed to write server ID to global env: %w", err)
		}
	} else {
		if err := config.SetEnvVar(config.TokenEnv, token); err != nil {
			return fmt.Errorf("failed to write token to %s: %w", config.EnvFileName, err)
		}
		if serverID != "" {
			if err := config.SetEnvVar(config.ServerIDEnv, serverID); err != nil {
				return fmt.Errorf("failed to write server ID to %s: %w", config.EnvFileName, err)
			}
		}
	}
	return nil
}

func readToken() (string, error) {
	if initFlags.token != "" {
		return strings.TrimSpace(initFlags.token), nil
	}
	if agentMode {
		return "", fmt.Errorf("no token provided; pass --token <token> (the interactive prompt is disabled by --agent)")
	}
	fmt.Println("Enter your Discord bot token:")
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func init() {
	initCmd.Flags().StringVar(&initFlags.token, "token", "", "Bot token (skips the interactive prompt; required with --agent)")
	initCmd.Flags().StringVar(&initFlags.configLocation, "config-location", "", "Config storage location: 'local' (cwd/server.json) or 'global' (~/.config/disc/servers/<id>/) (required with --agent)")
	initCmd.Flags().StringVar(&initFlags.serverID, "server-id", "", "Server ID to manage (required in global mode when bot is in multiple servers)")
}
