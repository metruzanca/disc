package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

// serverConfigFile is the declarative, combined server config: both roles and
// channels in one JSON document. Entries carry their server IDs so edits can
// reference resources the same way the server CLI does. Server records the ID
// of the server the config was pulled from, so push can warn when it is
// applied to a different server.
type serverConfigFile struct {
	Server   string          `json:"server_id,omitempty"`
	Roles    []roleExport    `json:"roles"`
	Channels []channelExport `json:"channels"`
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage server config as declarative JSON",
	Long: `Manage the server's roles and channels as a declarative JSON file
(server.json by default).

  disc config pull      - snapshot the live server into server.json
  disc config push      - apply server.json to the live server
  disc config role      - edit roles in server.json (add/update/delete)
  disc config channel   - edit channels in server.json (add/update/delete/move)`,
}

// readConfigFile loads a server config JSON file, giving a helpful error when
// it is missing so the user knows to pull first.
func readConfigFile(path string, cfg *serverConfigFile) error {
	if err := readJSONFile(path, cfg); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("config file %s not found; run 'disc config pull' first", path)
		}
		return err
	}
	return nil
}

// writeConfigFile marshals a server config to JSON and writes it to path,
// creating parent directories as needed.
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

// resolveConfigFile returns the config file path for the given server hint.
// The server hint may be "" if unknown; resolution will attempt to determine it
// from prefs or a single servers/ directory.
func resolveConfigFile(flagFile, serverHint string) (string, error) {
	if flagFile != "" {
		return flagFile, nil
	}

	if _, err := os.Stat("server.json"); err == nil {
		return "server.json", nil
	}

	prefs, err := config.LoadPrefs()
	if err != nil {
		return "", err
	}

	loc := prefs.ServerConfigLocation
	if envLoc := os.Getenv(config.ConfigLocationEnv); envLoc != "" {
		loc = envLoc
	}

	if loc == "global" || loc == "" {
		sid := serverHint
		if sid == "" {
			sid = os.Getenv(config.ServerIDEnv)
		}
		if sid == "" {
			sid = config.SingleServerDir()
		}
		if sid != "" {
			p, err := config.ServerConfigFile(sid)
			if err != nil {
				return "", err
			}
			return p, nil
		}
	}

	return "server.json", nil
}

// resolveServerHint returns a server ID hint from flag/env/single-dir.
func resolveServerHint(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if sid := os.Getenv(config.ServerIDEnv); sid != "" {
		return sid
	}
	return config.SingleServerDir()
}

var (
	configPullFlags = struct {
		server string
		file   string
	}{}
)

var configPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Snapshot the live server into server.json",
	Long: `Fetch the server's roles and channels and write them to server.json
(overwrites any existing file). Managed (integration/bot) roles and channel
kinds that are not managed are skipped.

Examples:
  disc config pull
  disc config pull --server 123456789
  disc config pull --file server.json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(configPullFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		s := client.Session()
		cfg := serverConfigFile{Server: serverID}
		if cfg.Roles, err = exportRoles(s, serverID); err != nil {
			return err
		}
		if cfg.Channels, err = exportChannels(s, serverID); err != nil {
			return err
		}

		out, err := resolveConfigFile(configPullFlags.file, serverID)
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

var (
	configPushFlags = struct {
		server        string
		file          string
		deleteMissing bool
		yes           bool
		dry           bool
	}{}
)

var configPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Apply server.json to the live server",
	Long: `Apply the roles and channels defined in server.json to the live server,
matching each entry by ID when present and by name otherwise.

An entry that already matches is left unchanged. Pass --delete-missing to also
delete resources on the server that are not in the file (a non-reversible
action). Managed roles, @everyone and system channels are never deleted. A dry
run is shown first; pass --yes to apply without prompting, or --dry to stop
after showing the plan.

Examples:
  disc config push
  disc config push --file server.json
  disc config push --delete-missing
  disc config push --dry`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(configPushFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		file, err := resolveConfigFile(configPushFlags.file, serverID)
		if err != nil {
			return err
		}
		var cfg serverConfigFile
		if err := readConfigFile(file, &cfg); err != nil {
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

		roleChanges, roleOps, err := planRoleImport(s, serverID, cfg.Roles, roles, roleIDByName, configPushFlags.deleteMissing)
		if err != nil {
			return err
		}
		channelChanges, channelOps, err := planChannelImport(s, guild, cfg.Channels, roleIDByName, configPushFlags.deleteMissing)
		if err != nil {
			return err
		}

		changes := append(roleChanges, channelChanges...)
		ops := append(roleOps, channelOps...)

		abs, err := filepath.Abs(file)
		if err != nil {
			abs = file
		}
		proceed, err := applyPlan(abs, changes, configPushFlags.yes, configPushFlags.dry)
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

var configRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Edit roles in the config file",
	Long:  "Add, update, or delete roles in the server config file.",
}

var configChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Edit channels in the config file",
	Long:  "Add, update, delete, or move channels in the server config file.",
}

func init() {
	configCmd.AddCommand(configPullCmd)
	configCmd.AddCommand(configPushCmd)
	configCmd.AddCommand(configRoleCmd)
	configCmd.AddCommand(configChannelCmd)

	configPullCmd.Flags().StringVar(&configPullFlags.server, "server", "", "Server ID (defaults to configured server)")
	configPullCmd.Flags().StringVar(&configPullFlags.file, "file", "", "Output file (defaults to server.json)")

	configPushCmd.Flags().StringVar(&configPushFlags.server, "server", "", "Server ID (defaults to configured server)")
	configPushCmd.Flags().StringVar(&configPushFlags.file, "file", "", "JSON file to push (defaults to server.json)")
	configPushCmd.Flags().BoolVar(&configPushFlags.deleteMissing, "delete-missing", false, "Delete resources on the server not present in the file (non-reversible)")
	configPushCmd.Flags().BoolVarP(&configPushFlags.yes, "yes", "y", false, "Skip confirmation prompt")
	configPushCmd.Flags().BoolVar(&configPushFlags.dry, "dry", false, "Show the plan without applying it")
}
