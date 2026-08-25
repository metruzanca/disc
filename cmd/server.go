package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/huh"
	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var errUserAborted = huh.ErrUserAborted

type serverConfigFile struct {
	Server   string          `json:"server_id,omitempty"`
	Bot      string          `json:"bot,omitempty"`
	Roles    []roleExport    `json:"roles"`
	Channels []channelExport `json:"channels"`
}

func readConfigFile(path string, cfg *serverConfigFile) error {
	if err := readJSONFile(path, cfg); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("config file %s not found; run 'disc server pull' first", path)
		}
		return err
	}
	return nil
}

func writeConfigFile(cfg *serverConfigFile, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create dir for %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func resolveConfigFile(fileFlag, serverHint string) (string, error) {
	if fileFlag != "" {
		return fileFlag, nil
	}

	if _, err := os.Stat("server.json"); err == nil {
		return "server.json", nil
	}

	bf, err := loadBotsForResolution()
	if err == nil && bf != nil {
		if serverHint != "" {
			if botID, ok := bf.BotManagingServer(serverHint); ok {
				if entry, ok := bf.GetServerEntry(botID, serverHint); ok && entry.LocalPath != "" {
					return filepath.Join(entry.LocalPath, "server.json"), nil
				}
				if p, err := config.ServerConfigFile(serverHint); err == nil {
					return p, nil
				}
			}
		}
		if bf.ActiveServer != "" {
			if botID := bf.ActiveBot; botID != "" {
				if entry, ok := bf.GetServerEntry(botID, bf.ActiveServer); ok && entry.LocalPath != "" {
					return filepath.Join(entry.LocalPath, "server.json"), nil
				}
				if p, err := config.ServerConfigFile(bf.ActiveServer); err == nil {
					return p, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no server configured; run 'disc server new' or use --file")
}

func loadBotsForResolution() (*botsFile, error) {
	bf, err := loadBots()
	if err != nil {
		return nil, err
	}
	if bf.Bots == nil || len(bf.Bots) == 0 {
		return nil, fmt.Errorf("no bots configured")
	}
	return bf, nil
}

type botsFile = config.BotsFile

func loadBots() (*botsFile, error) {
	return config.LoadBots()
}

func saveBots(bf *botsFile) error {
	return config.SaveBots(bf)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage Discord servers and bots",
	Long: `Commands for setting up, switching between, and managing Discord servers.

  disc server new     - Add a new bot and server to disc
  disc server switch  - Switch the active server
  disc server list    - List all managed servers
  disc server pull    - Snapshot the live server into the config file
  disc server push    - Apply the config file to the live server`,
}

var serverNewFlags = struct {
	token          string
	configLocation string
	serverID       string
}{}

var serverNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Add a new bot and server to disc",
	Long: `Interactively (or with flags) add a new bot token to disc, pick a server
the bot is in, choose a config location (local or global), and snapshot
the current server state into a server.json file.

In agent mode (--agent), --token and --config-location are required, and
--server-id is required when the bot is in multiple servers.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		tok := serverNewFlags.token
		if tok == "" {
			if agentMode {
				return fmt.Errorf("--token is required in agent mode")
			}
			fmt.Println("Enter your Discord bot token:")
			fmt.Print("> ")
			line, err := readLine()
			if err != nil {
				return err
			}
			tok = line
		}
		tok = trimTok(tok)
		if tok == "" {
			return fmt.Errorf("no token provided")
		}

		loc := serverNewFlags.configLocation
		if loc == "" {
			if agentMode {
				return fmt.Errorf("--config-location is required in agent mode (local or global)")
			}
			loc, err = promptConfigLocation()
			if err != nil {
				if errors.Is(err, errUserAborted) {
					util.Yellow.Println("Aborted.")
					return nil
				}
				return err
			}
		}
		if loc != "local" && loc != "global" {
			return fmt.Errorf("--config-location must be 'local' or 'global'")
		}

		client, err := discord.New(tok)
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}
		defer client.Close()
		if err := client.WaitReady(); err != nil {
			return err
		}

		botUser := client.User()
		botID := botUser.ID

		guilds := client.Guilds()

		var serverID string
		if serverNewFlags.serverID != "" {
			serverID = serverNewFlags.serverID
		} else if len(guilds) == 1 {
			serverID = guilds[0].ID
		} else if len(guilds) > 1 {
			if agentMode {
				return fmt.Errorf("--server-id is required when the bot is in multiple servers")
			}
			serverID, err = promptServerSelection(guilds)
			if err != nil {
				if errors.Is(err, errUserAborted) {
					util.Yellow.Println("Aborted.")
					return nil
				}
				return err
			}
		}

		bf, err := loadBots()
		if err != nil {
			return err
		}
		bf.UpsertBot(botID, tok)

		if serverID == "" {
			bf.SetActive(botID, "")
			if err := saveBots(bf); err != nil {
				return fmt.Errorf("failed to save bots.json: %w", err)
			}
			util.Green.Printf("Bot %s saved to disc.\n", botUser.Username)
			fmt.Println()
			fmt.Println("Invite your bot to a server using this link:")
			fmt.Println()
			fmt.Println(discord.InviteLink(botID))
			fmt.Println()
			util.Yellow.Println("Run 'disc server new' again once the bot is in a server.")
			return nil
		}

		var serverName string
		for _, g := range guilds {
			if g.ID == serverID {
				serverName = g.Name
				break
			}
		}

		var localPath string
		if loc == "global" {
			sd, err := config.ResolveServerDir(serverID, true)
			if err != nil {
				return err
			}
			localPath = ""
			_ = sd
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get cwd: %w", err)
			}
			localPath = cwd
		}

		bf.AddServerToBot(botID, serverID, serverName, localPath)
		bf.SetActive(botID, serverID)
		if err := saveBots(bf); err != nil {
			return fmt.Errorf("failed to save bots.json: %w", err)
		}

		var cfgPath string
		if loc == "global" {
			cfgPath, _ = config.ServerConfigFile(serverID)
		} else {
			cfgPath = filepath.Join(localPath, "server.json")
		}

		util.Green.Printf("Bot %s saved.\n", botUser.Username)
		util.Green.Printf("Server: %s\n", serverName)
		util.Green.Printf("Config location: %s\n", cfgPath)
		fmt.Println()
		fmt.Println("Invite your bot to a server using this link:")
		fmt.Println()
		fmt.Println(discord.InviteLink(botID))
		fmt.Println()

		s := client.Session()
		cfg := serverConfigFile{Server: serverID, Bot: botID}
		if cfg.Roles, err = exportRoles(s, serverID); err != nil {
			return err
		}
		if cfg.Channels, err = exportChannels(s, serverID); err != nil {
			return err
		}
		if err := writeConfigFile(&cfg, cfgPath); err != nil {
			return err
		}
		util.Green.Printf("Pulled config to %s (%d roles, %d channels)\n", cfgPath, len(cfg.Roles), len(cfg.Channels))

		_ = serverName
		return nil
	},
}

var serverSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the active server",
	Long: `Switch to a different managed server. Lists all servers from bots.json
with their location (local path or global). If a local server is selected,
prints a 'cd' hint.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetServer := serverSwitchFlags.server
		if targetServer != "" {
			bf, err := loadBots()
			if err != nil {
				return err
			}
			if bf.Bots == nil || len(bf.Bots) == 0 {
				return fmt.Errorf("no bots configured; run 'disc server new' first")
			}
			botID, ok := bf.BotManagingServer(targetServer)
			if !ok {
				return fmt.Errorf("server %s is not under management; run 'disc server new'", targetServer)
			}
			bf.SetActive(botID, targetServer)
			if err := saveBots(bf); err != nil {
				return err
			}
			util.Green.Printf("Switched to server %s.\n", targetServer)
			entry, _ := bf.GetServerEntry(botID, targetServer)
			if entry.LocalPath != "" {
				util.Green.Printf("cd %s\n", entry.LocalPath)
			}
			return nil
		}

		if agentMode {
			return fmt.Errorf("--server is required in agent mode")
		}

		bf, err := loadBots()
		if err != nil {
			return err
		}
		if bf.Bots == nil || len(bf.Bots) == 0 {
			return fmt.Errorf("no bots configured; run 'disc server new' first")
		}

		allServers := bf.AllServers()
		if len(allServers) == 0 {
			return fmt.Errorf("no servers under management; run 'disc server new' first")
		}

		options := make([]huh.Option[string], len(allServers))
		labels := make([]string, len(allServers))
		for i, s := range allServers {
			loc := "global"
			if s.Entry.LocalPath != "" {
				loc = s.Entry.LocalPath
			}
			active := ""
			if bf.ActiveBot == s.BotID && bf.ActiveServer == s.Server {
				active = " (active)"
			}
			label := fmt.Sprintf("%s / %s [%s]%s", s.Bot.ID, s.Entry.Name, loc, active)
			options[i] = huh.NewOption(label, s.Server)
			labels[i] = s.Server
		}

		var selected string
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a server").
				Options(options...).
				Value(&selected),
		))
		if err := form.Run(); err != nil {
			if errors.Is(err, errUserAborted) {
				util.Yellow.Println("Aborted.")
				return nil
			}
			return err
		}

		botID, _ := bf.BotManagingServer(selected)
		bf.SetActive(botID, selected)
		if err := saveBots(bf); err != nil {
			return err
		}

		entry, _ := bf.GetServerEntry(botID, selected)
		util.Green.Printf("Switched to server %s.\n", selected)
		if entry.LocalPath != "" {
			util.Green.Printf("cd %s\n", entry.LocalPath)
		}
		return nil
	},
}

var serverSwitchFlags = struct {
	server string
}{}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed servers",
	Long: `Prints a table of all bots and servers under disc management.
Use this in scripts that need to parse server IDs.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		bf, err := loadBots()
		if err != nil {
			return err
		}
		if bf.Bots == nil || len(bf.Bots) == 0 {
			util.Yellow.Println("No bots configured.")
			return nil
		}
		util.Bold.Println("Bots and servers:")
		for botID, bot := range bf.Bots {
			activeMarker := ""
			if bf.ActiveBot == botID {
				activeMarker = " [active bot]"
			}
			util.Cyan.Printf("  Bot: %s%s\n", botID, activeMarker)
			for sid, entry := range bot.Servers {
				loc := "~/.config/disc/servers/" + sid
				if entry.LocalPath != "" {
					loc = entry.LocalPath
				}
				activeSrv := ""
				if bf.ActiveBot == botID && bf.ActiveServer == sid {
					activeSrv = " [active]"
				}
				fmt.Printf("    %s / %s [%s]%s\n", sid, entry.Name, loc, activeSrv)
			}
		}
		return nil
	},
}

var serverPullFlags = struct {
	server string
	file   string
}{}

var serverPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Snapshot the live server into the config file",
	Long: `Fetch the server's roles and channels and write them to the config file
(overwrites any existing file). The config file is resolved via --file,
the local server.json, or the active server. Managed (integration/bot) roles
and channel kinds that are not managed are skipped.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(serverPullFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		botID := client.User().ID

		s := client.Session()
		cfg := serverConfigFile{Server: serverID, Bot: botID}
		if cfg.Roles, err = exportRoles(s, serverID); err != nil {
			return err
		}
		if cfg.Channels, err = exportChannels(s, serverID); err != nil {
			return err
		}

		out, err := resolveConfigFile(serverPullFlags.file, serverID)
		if err != nil {
			return err
		}
		if err := writeConfigFile(&cfg, out); err != nil {
			return err
		}
		util.Green.Printf("Pulled config to %s (%d roles, %d channels)\n", out, len(cfg.Roles), len(cfg.Channels))
		return nil
	},
}

var serverPushFlags = struct {
	file          string
	deleteMissing bool
	yes           bool
	dry           bool
}{}

var serverPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Apply the config file to the live server",
	Long: `Apply the roles and channels defined in the config file to the live server,
matching each entry by ID when present and by name otherwise.

An entry that already matches is left unchanged. Pass --delete-missing to also
delete resources on the server not in the file (non-reversible). Managed roles,
@everyone and system channels are never deleted. A dry run is shown first;
pass --yes to apply without prompting, or --dry to stop after the plan.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		file, botID, err := resolveConfigFileForServer("", serverPushFlags.file)
		if err != nil {
			return err
		}
		var cfg serverConfigFile
		if err := readConfigFile(file, &cfg); err != nil {
			return err
		}

		serverID := cfg.Server
		if serverID == "" {
			return fmt.Errorf("config file has no server_id; run 'disc server pull' first")
		}

		bf, err := loadBots()
		if err != nil {
			return err
		}
		botEntry, _ := bf.GetBot(botID)
		token := botEntry.Token
		if token == "" {
			return fmt.Errorf("no token found for bot %s; run 'disc server new'", botID)
		}

		client, err := discord.New(token)
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}
		defer client.Close()
		if err := client.WaitReady(); err != nil {
			return err
		}

		s := client.Session()
		guild, err := s.Guild(serverID)
		if err != nil {
			return fmt.Errorf("failed to load server: %w", err)
		}

		if cfg.Server != "" && cfg.Server != serverID {
			util.Red.Printf("WARNING: %s was pulled from server %s, but you are pushing to server %s (%s).\n", file, cfg.Server, serverID, guild.Name)
			util.Red.Println("If this isn't intentional (e.g. cloning to a new server), abort now.")
			fmt.Println()
		}

		roles, err := s.GuildRoles(serverID)
		if err != nil {
			return fmt.Errorf("failed to list roles: %w", err)
		}

		roleIDByName := map[string]string{}
		for _, r := range roles {
			if r.ID == serverID {
				roleIDByName["@everyone"] = r.ID
			} else {
				roleIDByName[r.Name] = r.ID
			}
		}

		roleChanges, roleOps, err := planRoleImport(s, serverID, cfg.Roles, roles, roleIDByName, serverPushFlags.deleteMissing)
		if err != nil {
			return err
		}
		channelChanges, channelOps, err := planChannelImport(s, guild, cfg.Channels, roleIDByName, serverPushFlags.deleteMissing)
		if err != nil {
			return err
		}

		changes := append(roleChanges, channelChanges...)
		ops := append(roleOps, channelOps...)

		abs, err := filepath.Abs(file)
		if err != nil {
			abs = file
		}
		proceed, err := applyPlan(abs, changes, serverPushFlags.yes, serverPushFlags.dry)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
		for _, op := range ops {
			if err := op(); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	serverCmd.AddCommand(serverNewCmd)
	serverCmd.AddCommand(serverSwitchCmd)
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverPullCmd)
	serverCmd.AddCommand(serverPushCmd)

	serverNewCmd.Flags().StringVar(&serverNewFlags.token, "token", "", "Bot token (skips the interactive prompt; required with --agent)")
	serverNewCmd.Flags().StringVar(&serverNewFlags.configLocation, "config-location", "", "Config storage location: 'local' (cwd/server.json) or 'global' (~/.config/disc/servers/<id>/) (required with --agent)")
	serverNewCmd.Flags().StringVar(&serverNewFlags.serverID, "server-id", "", "Server ID to manage (required when the bot is in multiple servers)")

	serverSwitchCmd.Flags().StringVar(&serverSwitchFlags.server, "server", "", "Server ID to switch to (non-interactive)")

	serverPullCmd.Flags().StringVar(&serverPullFlags.server, "server", "", "Server ID (defaults to active server)")
	serverPullCmd.Flags().StringVar(&serverPullFlags.file, "file", "", "Output file (defaults to resolved server.json)")

	serverPushCmd.Flags().StringVar(&serverPushFlags.file, "file", "", "JSON file to push (defaults to resolved server.json)")
	serverPushCmd.Flags().BoolVar(&serverPushFlags.deleteMissing, "delete-missing", false, "Delete resources on the server not present in the file (non-reversible)")
	serverPushCmd.Flags().BoolVarP(&serverPushFlags.yes, "yes", "y", false, "Skip confirmation prompt")
	serverPushCmd.Flags().BoolVar(&serverPushFlags.dry, "dry", false, "Show the plan without applying it")
}

func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func trimTok(s string) string {
	return strings.TrimSpace(s)
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
