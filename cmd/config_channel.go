package cmd

import (
	"fmt"

	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var (
	configChannelAddFlags = struct {
		file     string
		name     string
		typ      string
		topic    string
		nsfw     bool
		category string
		position int
		allow    []string
		deny     []string
	}{}
	configChannelUpFlags = struct {
		file        string
		channel     string
		name        string
		topic       string
		nsfw        bool
		category    string
		allow       []string
		deny        []string
		noOverwrite []string
	}{}
	configChannelDelFlags = struct {
		file    string
		channel string
	}{}
	configChannelMoveFlags = struct {
		file     string
		channel  string
		position int
		category string
	}{}
)

// findChannelInConfig returns the index of a channel by ID, or -1.
func findChannelInConfig(cfg *serverConfigFile, id string) int {
	for i := range cfg.Channels {
		if cfg.Channels[i].ID == id {
			return i
		}
	}
	return -1
}

// applyConfigOverwrites merges --allow/--deny values (each "Role:Perm,Perm")
// into a channel's config overwrites, replacing the provided side for each
// targeted role and preserving the other side.
func applyConfigOverwrites(ce *channelExport, allows, denies []string) error {
	targets, err := collectPermTargets(allows, denies)
	if err != nil {
		return err
	}
	idx := map[string]int{}
	for i, o := range ce.Overwrites {
		idx[o.Role] = i
	}
	for role, t := range targets {
		i, exists := idx[role]
		ow := permissionOverwriteExport{Role: role}
		if exists {
			ow = ce.Overwrites[i]
		}
		if t.allowChanged {
			ow.Allow = t.allow
		}
		if t.denyChanged {
			ow.Deny = t.deny
		}
		if exists {
			ce.Overwrites[i] = ow
		} else {
			ce.Overwrites = append(ce.Overwrites, ow)
		}
	}
	return nil
}

// removeConfigOverwrite removes the given roles' overwrites from a channel.
func removeConfigOverwrite(ce *channelExport, roles []string) {
	removed := map[string]bool{}
	for _, r := range roles {
		removed[r] = true
	}
	keep := ce.Overwrites[:0]
	for _, o := range ce.Overwrites {
		if !removed[o.Role] {
			keep = append(keep, o)
		}
	}
	ce.Overwrites = keep
}

var configChannelAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a channel to the config",
	Long: `Add a channel to the config file. The channel is not created on the
server until 'disc config push' runs. If a channel with the same name already
exists in the config, an error is returned.

Examples:
  disc config channel add --name general
  disc config channel add --name voice-chat --type voice
  disc config channel add --name help --category Staff`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := loadConfigFileForEdit(configChannelAddFlags.file)
		if err != nil {
			return err
		}
		if configChannelAddFlags.name == "" {
			return fmt.Errorf("--name is required")
		}
		for _, c := range cfg.Channels {
			if c.Name == configChannelAddFlags.name {
				return fmt.Errorf("channel '%s' already exists in config; use 'disc config channel update' instead", c.Name)
			}
		}

		typ := "text"
		if configChannelAddFlags.typ != "" {
			typ = configChannelAddFlags.typ
		}
		if _, ok := channelTypeFromName(typ); !ok {
			return fmt.Errorf("unsupported channel type '%s'", typ)
		}

		ce := channelExport{
			Name:     configChannelAddFlags.name,
			Type:     typ,
			Topic:    configChannelAddFlags.topic,
			NSFW:     configChannelAddFlags.nsfw,
			Parent:   configChannelAddFlags.category,
			Position: configChannelAddFlags.position,
		}
		if len(configChannelAddFlags.allow) > 0 || len(configChannelAddFlags.deny) > 0 {
			if err := applyConfigOverwrites(&ce, configChannelAddFlags.allow, configChannelAddFlags.deny); err != nil {
				return err
			}
		}
		cfg.Channels = append(cfg.Channels, ce)
		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Added channel '%s' to %s\n", ce.Name, path)
		return nil
	},
}

var configChannelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a channel in the config",
	Long: `Update a channel in the config file by its ID. The change is applied
to the server on the next 'disc config push'. Use --category none to remove a
channel from its category.

Examples:
  disc config channel update --channel 123456789 --name new-name
  disc config channel update --channel 123456789 --topic "Welcome to the channel"
  disc config channel update --channel 123456789 --category Staff
  disc config channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
  disc config channel update --channel 123456789 --no-overwrite Moderator`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := loadConfigFileForEdit(configChannelUpFlags.file)
		if err != nil {
			return err
		}
		i := findChannelInConfig(cfg, configChannelUpFlags.channel)
		if i == -1 {
			return fmt.Errorf("channel with ID '%s' not found in config", configChannelUpFlags.channel)
		}

		ce := &cfg.Channels[i]
		if cmd.Flags().Changed("name") {
			ce.Name = configChannelUpFlags.name
		}
		if cmd.Flags().Changed("topic") {
			ce.Topic = configChannelUpFlags.topic
		}
		if cmd.Flags().Changed("nsfw") {
			ce.NSFW = configChannelUpFlags.nsfw
		}
		if cmd.Flags().Changed("category") {
			if configChannelUpFlags.category == "none" {
				ce.Parent = ""
			} else {
				ce.Parent = configChannelUpFlags.category
			}
		}
		if len(configChannelUpFlags.allow) > 0 || len(configChannelUpFlags.deny) > 0 {
			if err := applyConfigOverwrites(ce, configChannelUpFlags.allow, configChannelUpFlags.deny); err != nil {
				return err
			}
		}
		if len(configChannelUpFlags.noOverwrite) > 0 {
			removeConfigOverwrite(ce, configChannelUpFlags.noOverwrite)
		}

		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Updated channel '%s' in %s\n", ce.Name, path)
		return nil
	},
}

var configChannelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a channel from the config",
	Long: `Remove a channel from the config file by its ID. The channel is
removed from the server on the next 'disc config push'.

Example:
  disc config channel delete --channel 123456789`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := loadConfigFileForEdit(configChannelDelFlags.file)
		if err != nil {
			return err
		}
		i := findChannelInConfig(cfg, configChannelDelFlags.channel)
		if i == -1 {
			return fmt.Errorf("channel with ID '%s' not found in config", configChannelDelFlags.channel)
		}
		name := cfg.Channels[i].Name
		cfg.Channels = append(cfg.Channels[:i], cfg.Channels[i+1:]...)
		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Removed channel '%s' from %s\n", name, path)
		return nil
	},
}

var configChannelMoveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a channel within the config",
	Long: `Move a channel to a new position or category in the config file. The
change is applied to the server on the next 'disc config push'. Use
--category none to remove a channel from its category.

Examples:
  disc config channel move --channel 123456789 --position 0
  disc config channel move --channel 123456789 --category Staff
  disc config channel move --channel 123456789 --category Staff --position 3`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := loadConfigFileForEdit(configChannelMoveFlags.file)
		if err != nil {
			return err
		}
		i := findChannelInConfig(cfg, configChannelMoveFlags.channel)
		if i == -1 {
			return fmt.Errorf("channel with ID '%s' not found in config", configChannelMoveFlags.channel)
		}

		ce := &cfg.Channels[i]
		if cmd.Flags().Changed("position") {
			ce.Position = configChannelMoveFlags.position
		}
		if cmd.Flags().Changed("category") {
			if configChannelMoveFlags.category == "none" {
				ce.Parent = ""
			} else {
				ce.Parent = configChannelMoveFlags.category
			}
		}

		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Moved channel '%s' in %s\n", ce.Name, path)
		return nil
	},
}

func init() {
	configChannelCmd.AddCommand(configChannelAddCmd)
	configChannelCmd.AddCommand(configChannelUpdateCmd)
	configChannelCmd.AddCommand(configChannelDeleteCmd)
	configChannelCmd.AddCommand(configChannelMoveCmd)

	configChannelAddCmd.Flags().StringVar(&configChannelAddFlags.file, "file", "", "Config file (defaults to server.json)")
	configChannelAddCmd.Flags().StringVar(&configChannelAddFlags.name, "name", "", "Channel name (required)")
	configChannelAddCmd.Flags().StringVar(&configChannelAddFlags.typ, "type", "text", "Channel type: text, voice, category, news, stage, forum")
	configChannelAddCmd.Flags().StringVar(&configChannelAddFlags.topic, "topic", "", "Channel topic (text channels only)")
	configChannelAddCmd.Flags().BoolVar(&configChannelAddFlags.nsfw, "nsfw", false, "Mark channel as NSFW")
	configChannelAddCmd.Flags().StringVar(&configChannelAddFlags.category, "category", "", "Category name to place channel under")
	configChannelAddCmd.Flags().IntVar(&configChannelAddFlags.position, "position", 0, "Position within the category")
	configChannelAddCmd.Flags().StringSliceVar(&configChannelAddFlags.allow, "allow", nil, "Permission overwrite to allow, as 'Role:Perm,Perm' (repeatable)")
	configChannelAddCmd.Flags().StringSliceVar(&configChannelAddFlags.deny, "deny", nil, "Permission overwrite to deny, as 'Role:Perm,Perm' (repeatable)")

	configChannelUpdateCmd.Flags().StringVar(&configChannelUpFlags.file, "file", "", "Config file (defaults to server.json)")
	configChannelUpdateCmd.Flags().StringVar(&configChannelUpFlags.channel, "channel", "", "Channel ID to update (required)")
	configChannelUpdateCmd.Flags().StringVar(&configChannelUpFlags.name, "name", "", "New channel name")
	configChannelUpdateCmd.Flags().StringVar(&configChannelUpFlags.topic, "topic", "", "Channel topic (text channels only)")
	configChannelUpdateCmd.Flags().BoolVar(&configChannelUpFlags.nsfw, "nsfw", false, "Mark channel as NSFW")
	configChannelUpdateCmd.Flags().StringVar(&configChannelUpFlags.category, "category", "", "Category name to move channel to (use 'none' to remove from category)")
	configChannelUpdateCmd.Flags().StringSliceVar(&configChannelUpFlags.allow, "allow", nil, "Permission overwrite to allow, as 'Role:Perm,Perm' (repeatable)")
	configChannelUpdateCmd.Flags().StringSliceVar(&configChannelUpFlags.deny, "deny", nil, "Permission overwrite to deny, as 'Role:Perm,Perm' (repeatable)")
	configChannelUpdateCmd.Flags().StringSliceVar(&configChannelUpFlags.noOverwrite, "no-overwrite", nil, "Remove this role's permission overwrite (repeatable)")

	configChannelDeleteCmd.Flags().StringVar(&configChannelDelFlags.file, "file", "", "Config file (defaults to server.json)")
	configChannelDeleteCmd.Flags().StringVar(&configChannelDelFlags.channel, "channel", "", "Channel ID to delete (required)")

	configChannelMoveCmd.Flags().StringVar(&configChannelMoveFlags.file, "file", "", "Config file (defaults to server.json)")
	configChannelMoveCmd.Flags().StringVar(&configChannelMoveFlags.channel, "channel", "", "Channel ID to move (required)")
	configChannelMoveCmd.Flags().IntVar(&configChannelMoveFlags.position, "position", 0, "New position within the category (0-indexed)")
	configChannelMoveCmd.Flags().StringVar(&configChannelMoveFlags.category, "category", "", "Category name to move channel to (use 'none' to remove from category)")
}
