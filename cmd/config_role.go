package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

// configRoleEditFlags are the value flags shared by role add/update.
type configRoleEditFlags struct {
	file        string
	server      string
	name        string
	color       string
	hoist       bool
	mentionable bool
	permissions string
	position    int
}

// configRoleUpdateFlags adds the --role selector to the shared edit flags.
type configRoleUpdateFlags struct {
	configRoleEditFlags
	role string
}

var (
	configRoleAddFlags = configRoleEditFlags{}
	configRoleUpFlags  = configRoleUpdateFlags{}
	configRoleDelFlags = struct {
		file   string
		server string
		role   string
	}{}
)

// splitPermissions splits and validates a comma-separated permission list.
func splitPermissions(s string) ([]string, error) {
	permBits := permissionBitsByName()
	var perms []string
	for _, part := range strings.Split(s, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, ok := permBits[name]; !ok {
			return nil, fmt.Errorf("unknown permission '%s'; see 'disc role perm list'", name)
		}
		perms = append(perms, name)
	}
	return perms, nil
}

// findRoleInConfig returns the index of a role by ID, or -1.
func findRoleInConfig(cfg *serverConfigFile, id string) int {
	for i := range cfg.Roles {
		if cfg.Roles[i].ID == id {
			return i
		}
	}
	return -1
}

// loadConfigFileForEdit loads the config file for an edit command. If create
// is true and the file does not exist, it creates an empty config skeleton at
// the resolved path.
func loadConfigFileForEdit(file, serverHint string, create bool) (*serverConfigFile, string, error) {
	path, err := resolveConfigFile(file, serverHint)
	if err != nil {
		return nil, "", err
	}
	var cfg serverConfigFile
	if err := readConfigFile(path, &cfg); err != nil {
		if create && os.IsNotExist(err) {
			if err := writeConfigFile(&cfg, path); err != nil {
				return nil, "", err
			}
			return &cfg, path, nil
		}
		return nil, "", err
	}
	return &cfg, path, nil
}

var configRoleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a role to the config",
	Long: `Add a role to the config file. The role is not created on the server
until 'disc config push' runs. If a role with the same name already exists in
the config, an error is returned.

Examples:
  disc config role add --name Moderator
  disc config role add --name VIP --color FF0000 --hoist
  disc config role add --name Helper --permissions "Send Messages,Add Reactions"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		hint := resolveServerHint(configRoleAddFlags.server)
		cfg, path, err := loadConfigFileForEdit(configRoleAddFlags.file, hint, true)
		if err != nil {
			return err
		}
		if configRoleAddFlags.name == "" {
			return fmt.Errorf("--name is required")
		}
		for _, r := range cfg.Roles {
			if r.Name == configRoleAddFlags.name {
				return fmt.Errorf("role '%s' already exists in config; use 'disc config role update' instead", r.Name)
			}
		}

		re := roleExport{Name: configRoleAddFlags.name, Hoist: configRoleAddFlags.hoist, Mentionable: configRoleAddFlags.mentionable}
		if configRoleAddFlags.color != "" {
			if _, err := parseHexColor(configRoleAddFlags.color); err != nil {
				return err
			}
			re.Color = strings.ToUpper(configRoleAddFlags.color)
		}
		if cmd.Flags().Changed("permissions") {
			perms, err := splitPermissions(configRoleAddFlags.permissions)
			if err != nil {
				return err
			}
			re.Permissions = perms
		}
		cfg.Roles = append(cfg.Roles, re)
		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Added role '%s' to %s\n", re.Name, path)
		return nil
	},
}

var configRoleUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a role in the config",
	Long: `Update a role in the config file by its ID. The change is applied to
the server on the next 'disc config push'.

Examples:
  disc config role update --role 987654321 --name "New Name"
  disc config role update --role 987654321 --color 00FF00
  disc config role update --role 987654321 --permissions "Kick Members,Ban Members"
  disc config role update --role 987654321 --position 3`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		hint := resolveServerHint(configRoleUpFlags.server)
		cfg, path, err := loadConfigFileForEdit(configRoleUpFlags.file, hint, false)
		if err != nil {
			return err
		}
		i := findRoleInConfig(cfg, configRoleUpFlags.role)
		if i == -1 {
			return fmt.Errorf("role with ID '%s' not found in config", configRoleUpFlags.role)
		}

		re := &cfg.Roles[i]
		if cmd.Flags().Changed("name") {
			re.Name = configRoleUpFlags.name
		}
		if cmd.Flags().Changed("color") {
			if configRoleUpFlags.color == "" {
				re.Color = ""
			} else {
				if _, err := parseHexColor(configRoleUpFlags.color); err != nil {
					return err
				}
				re.Color = strings.ToUpper(configRoleUpFlags.color)
			}
		}
		if cmd.Flags().Changed("hoist") {
			re.Hoist = configRoleUpFlags.hoist
		}
		if cmd.Flags().Changed("mentionable") {
			re.Mentionable = configRoleUpFlags.mentionable
		}
		if cmd.Flags().Changed("permissions") {
			perms, err := splitPermissions(configRoleUpFlags.permissions)
			if err != nil {
				return err
			}
			re.Permissions = perms
		}
		if cmd.Flags().Changed("position") {
			re.Position = configRoleUpFlags.position
		}

		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Updated role '%s' in %s\n", re.Name, path)
		return nil
	},
}

var configRoleDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a role from the config",
	Long: `Remove a role from the config file by its ID. The role is removed
from the server on the next 'disc config push'.

Example:
  disc config role delete --role 987654321`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		hint := resolveServerHint(configRoleDelFlags.server)
		cfg, path, err := loadConfigFileForEdit(configRoleDelFlags.file, hint, false)
		if err != nil {
			return err
		}
		i := findRoleInConfig(cfg, configRoleDelFlags.role)
		if i == -1 {
			return fmt.Errorf("role with ID '%s' not found in config", configRoleDelFlags.role)
		}
		name := cfg.Roles[i].Name
		cfg.Roles = append(cfg.Roles[:i], cfg.Roles[i+1:]...)
		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Removed role '%s' from %s\n", name, path)
		return nil
	},
}

func init() {
	configRoleCmd.AddCommand(configRoleAddCmd)
	configRoleCmd.AddCommand(configRoleUpdateCmd)
	configRoleCmd.AddCommand(configRoleDeleteCmd)

	configRoleAddCmd.Flags().StringVar(&configRoleAddFlags.file, "file", "", "Config file (defaults to resolved location)")
	configRoleAddCmd.Flags().StringVar(&configRoleAddFlags.server, "server", "", "Server ID (for global config location)")
	configRoleAddCmd.Flags().StringVar(&configRoleAddFlags.name, "name", "", "Role name (required)")
	configRoleAddCmd.Flags().StringVar(&configRoleAddFlags.color, "color", "", "Role color in hex (e.g., FF0000 for red)")
	configRoleAddCmd.Flags().BoolVar(&configRoleAddFlags.hoist, "hoist", false, "Display role separately in member list")
	configRoleAddCmd.Flags().BoolVar(&configRoleAddFlags.mentionable, "mentionable", false, "Allow anyone to mention this role")
	configRoleAddCmd.Flags().StringVar(&configRoleAddFlags.permissions, "permissions", "", "Comma-separated permission names; see 'disc role perm list'")

	configRoleUpdateCmd.Flags().StringVar(&configRoleUpFlags.file, "file", "", "Config file (defaults to resolved location)")
	configRoleUpdateCmd.Flags().StringVar(&configRoleUpFlags.server, "server", "", "Server ID (for global config location)")
	configRoleUpdateCmd.Flags().StringVar(&configRoleUpFlags.role, "role", "", "Role ID to update (required)")
	configRoleUpdateCmd.Flags().StringVar(&configRoleUpFlags.name, "name", "", "New role name")
	configRoleUpdateCmd.Flags().StringVar(&configRoleUpFlags.color, "color", "", "Role color in hex (e.g., FF0000 for red)")
	configRoleUpdateCmd.Flags().BoolVar(&configRoleUpFlags.hoist, "hoist", false, "Display role separately in member list")
	configRoleUpdateCmd.Flags().BoolVar(&configRoleUpFlags.mentionable, "mentionable", false, "Allow anyone to mention this role")
	configRoleUpdateCmd.Flags().StringVar(&configRoleUpFlags.permissions, "permissions", "", "Comma-separated permission names; see 'disc role perm list'")
	configRoleUpdateCmd.Flags().IntVar(&configRoleUpFlags.position, "position", 0, "Role position (hierarchy rank; higher is above)")

	configRoleDeleteCmd.Flags().StringVar(&configRoleDelFlags.file, "file", "", "Config file (defaults to resolved location)")
	configRoleDeleteCmd.Flags().StringVar(&configRoleDelFlags.server, "server", "", "Server ID (for global config location)")
	configRoleDeleteCmd.Flags().StringVar(&configRoleDelFlags.role, "role", "", "Role ID to delete (required)")
}
