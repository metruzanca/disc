package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage Discord roles",
	Long:  "Commands for listing and managing Discord roles in a server.",
}

var (
	roleListFlags = struct{ server string }{}
)

var roleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List roles in a server",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(roleListFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		roles, err := client.Session().GuildRoles(serverID)
		if err != nil {
			return fmt.Errorf("failed to list roles: %w", err)
		}

		// Sort by hierarchy (position descending, highest first).
		sort.Slice(roles, func(i, j int) bool { return roles[i].Position > roles[j].Position })

		util.Bold.Println("Roles (sorted by hierarchy, highest first):")
		fmt.Println()
		for _, r := range roles {
			managed := ""
			if r.Managed {
				managed = " [managed]"
			}
			fmt.Printf("  %s%s ", r.Name, managed)
			util.Dim.Printf("(ID: %s)\n", r.ID)
		}
		return nil
	},
}

var (
	roleShowFlags = struct {
		server string
		role   string
	}{}
)

var roleShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show role details",
	Example: `  disc role show --server 123456789 --role 987654321`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if roleShowFlags.role == "" {
			return fmt.Errorf("--role is required")
		}
		client, serverID, err := newClientAndServer(roleShowFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		roles, err := client.Session().GuildRoles(serverID)
		if err != nil {
			return fmt.Errorf("failed to load roles: %w", err)
		}

		var role *discordgo.Role
		for _, r := range roles {
			if r.ID == roleShowFlags.role {
				role = r
				break
			}
		}
		if role == nil {
			return fmt.Errorf("role %s not found", roleShowFlags.role)
		}

		util.Bold.Printf("Role: %s\n", role.Name)
		util.Cyan.Printf("ID: %s\n", role.ID)
		util.Cyan.Printf("Position: %d\n", role.Position)
		colorDesc := "(default)"
		if role.Color != 0 {
			colorDesc = fmt.Sprintf("#%06X", role.Color)
		}
		util.Cyan.Printf("Color: %s\n", colorDesc)
		util.Cyan.Printf("Hoisted: %t\n", role.Hoist)
		util.Cyan.Printf("Mentionable: %t\n", role.Mentionable)
		util.Cyan.Printf("Managed: %t\n", role.Managed)
		fmt.Println()
		util.Bold.Println("Permissions:")
		for _, p := range permissionNames() {
			if role.Permissions&p.bit != 0 {
				util.Green.Printf("  + %s\n", p.name)
			}
		}
		return nil
	},
}

// permissionNames returns a display name for each named Discord permission,
// ordered from most to least privileged.
func permissionNames() []struct {
	name string
	bit  int64
} {
	return []struct {
		name string
		bit  int64
	}{
		{"Administrator", discordgo.PermissionAdministrator},
		{"Manage Server", discordgo.PermissionManageServer},
		{"View Audit Log", discordgo.PermissionViewAuditLogs},
		{"Manage Roles", discordgo.PermissionManageRoles},
		{"Manage Channels", discordgo.PermissionManageChannels},
		{"Manage Webhooks", discordgo.PermissionManageWebhooks},
		{"Create Invite", discordgo.PermissionCreateInstantInvite},
		{"Manage Events", discordgo.PermissionManageEvents},
		{"View Channels", discordgo.PermissionViewChannel},
		{"Send Messages", discordgo.PermissionSendMessages},
		{"Manage Messages", discordgo.PermissionManageMessages},
		{"Read Message History", discordgo.PermissionReadMessageHistory},
		{"Mention Everyone", discordgo.PermissionMentionEveryone},
		{"Kick Members", discordgo.PermissionKickMembers},
		{"Ban Members", discordgo.PermissionBanMembers},
		{"Manage Nicknames", discordgo.PermissionManageNicknames},
		{"Change Nickname", discordgo.PermissionChangeNickname},
		{"Manage Guild Expressions", discordgo.PermissionManageGuildExpressions},
		{"View Channel", discordgo.PermissionViewChannel},
		{"Send Messages In Threads", discordgo.PermissionSendMessagesInThreads},
		{"Manage Threads", discordgo.PermissionManageThreads},
		{"Embed Links", discordgo.PermissionEmbedLinks},
		{"Attach Files", discordgo.PermissionAttachFiles},
		{"Add Reactions", discordgo.PermissionAddReactions},
		{"Use External Emojis", discordgo.PermissionUseExternalEmojis},
		{"Connect", discordgo.PermissionVoiceConnect},
		{"Speak", discordgo.PermissionVoiceSpeak},
		{"Mute Members", discordgo.PermissionVoiceMuteMembers},
		{"Deafen Members", discordgo.PermissionVoiceDeafenMembers},
		{"Move Members", discordgo.PermissionVoiceMoveMembers},
		{"Use Voice Activity", discordgo.PermissionVoiceUseVAD},
		{"Priority Speaker", discordgo.PermissionVoicePrioritySpeaker},
	}
}

var (
	roleAddFlags = struct {
		server      string
		name        string
		color       string
		hoist       bool
		mentionable bool
		yes         bool
	}{}
)

var roleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new role",
	Long: `Create a new role in a Discord server.

Examples:
  disc role add --server 123456789 --name "Moderator"
  disc role add --server 123456789 --name "VIP" --color FF0000 --hoist
  disc role add --server 123456789 --name "Helper" --mentionable`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(roleAddFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()
		if roleAddFlags.name == "" {
			return fmt.Errorf("--name is required")
		}

		if !util.Confirm(fmt.Sprintf("Create role '%s' in server %s?", roleAddFlags.name, serverID), roleAddFlags.yes) {
			util.Yellow.Println("Aborted.")
			return nil
		}

		params := &discordgo.RoleParams{
			Name:        roleAddFlags.name,
			Hoist:       &roleAddFlags.hoist,
			Mentionable: &roleAddFlags.mentionable,
		}
		if roleAddFlags.color != "" {
			c, err := parseHexColor(roleAddFlags.color)
			if err != nil {
				return err
			}
			params.Color = &c
		}

		role, err := client.Session().GuildRoleCreate(serverID, params)
		if err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}
		util.Green.Printf("Created role %s (%s)\n", role.Name, role.ID)
		return nil
	},
}

var (
	roleUpdateFlags = struct {
		server      string
		role        string
		name        string
		color       string
		hoist       bool
		mentionable bool
		yes         bool
	}{}
)

var roleUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a role",
	Long: `Update an existing role's properties.

Examples:
  disc role update --server 123456789 --role 987654321 --name "New Name"
  disc role update --server 123456789 --role 987654321 --color 00FF00
  disc role update --server 123456789 --role 987654321 --hoist=false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if roleUpdateFlags.role == "" {
			return fmt.Errorf("--role is required")
		}
		client, serverID, err := newClientAndServer(roleUpdateFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		if !util.Confirm(fmt.Sprintf("Update role %s?", roleUpdateFlags.role), roleUpdateFlags.yes) {
			util.Yellow.Println("Aborted.")
			return nil
		}

		params := &discordgo.RoleParams{}
		if roleUpdateFlags.name != "" {
			params.Name = roleUpdateFlags.name
		}
		if roleUpdateFlags.color != "" {
			c, err := parseHexColor(roleUpdateFlags.color)
			if err != nil {
				return err
			}
			params.Color = &c
		}
		if cmd.Flags().Changed("hoist") {
			params.Hoist = &roleUpdateFlags.hoist
		}
		if cmd.Flags().Changed("mentionable") {
			params.Mentionable = &roleUpdateFlags.mentionable
		}

		role, err := client.Session().GuildRoleEdit(serverID, roleUpdateFlags.role, params)
		if err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}
		util.Green.Printf("Updated role %s (%s)\n", role.Name, role.ID)
		return nil
	},
}

var (
	roleDeleteFlags = struct {
		server string
		role   string
		yes    bool
	}{}
)

var roleDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a role",
	RunE: func(cmd *cobra.Command, args []string) error {
		if roleDeleteFlags.role == "" {
			return fmt.Errorf("--role is required")
		}
		client, serverID, err := newClientAndServer(roleDeleteFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		if !util.Confirm(fmt.Sprintf("Delete role %s?", roleDeleteFlags.role), roleDeleteFlags.yes) {
			util.Yellow.Println("Aborted.")
			return nil
		}

		if err := client.Session().GuildRoleDelete(serverID, roleDeleteFlags.role); err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}
		util.Green.Printf("Deleted role %s\n", roleDeleteFlags.role)
		return nil
	},
}

func parseHexColor(s string) (int, error) {
	v, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid color '%s': must be hex like FF0000", s)
	}
	return int(v), nil
}

func init() {
	roleCmd.AddCommand(roleListCmd)
	roleCmd.AddCommand(roleShowCmd)
	roleCmd.AddCommand(roleAddCmd)
	roleCmd.AddCommand(roleUpdateCmd)
	roleCmd.AddCommand(roleDeleteCmd)

	roleListCmd.Flags().StringVar(&roleListFlags.server, "server", "", "Server ID (defaults to configured server)")

	roleShowCmd.Flags().StringVar(&roleShowFlags.server, "server", "", "Server ID (defaults to configured server)")
	roleShowCmd.Flags().StringVar(&roleShowFlags.role, "role", "", "Role ID to show details for (required)")

	roleAddCmd.Flags().StringVar(&roleAddFlags.server, "server", "", "Server ID (defaults to configured server)")
	roleAddCmd.Flags().StringVar(&roleAddFlags.name, "name", "", "Role name (required)")
	roleAddCmd.Flags().StringVar(&roleAddFlags.color, "color", "", "Role color in hex (e.g., FF0000 for red)")
	roleAddCmd.Flags().BoolVar(&roleAddFlags.hoist, "hoist", false, "Display role separately in member list")
	roleAddCmd.Flags().BoolVar(&roleAddFlags.mentionable, "mentionable", false, "Allow anyone to mention this role")
	roleAddCmd.Flags().BoolVarP(&roleAddFlags.yes, "yes", "y", false, "Skip confirmation prompt")

	roleUpdateCmd.Flags().StringVar(&roleUpdateFlags.server, "server", "", "Server ID (defaults to configured server)")
	roleUpdateCmd.Flags().StringVar(&roleUpdateFlags.role, "role", "", "Role ID to update (required)")
	roleUpdateCmd.Flags().StringVar(&roleUpdateFlags.name, "name", "", "New role name")
	roleUpdateCmd.Flags().StringVar(&roleUpdateFlags.color, "color", "", "Role color in hex (e.g., FF0000 for red)")
	roleUpdateCmd.Flags().BoolVar(&roleUpdateFlags.hoist, "hoist", false, "Display role separately in member list")
	roleUpdateCmd.Flags().BoolVar(&roleUpdateFlags.mentionable, "mentionable", false, "Allow anyone to mention this role")
	roleUpdateCmd.Flags().BoolVarP(&roleUpdateFlags.yes, "yes", "y", false, "Skip confirmation prompt")

	roleDeleteCmd.Flags().StringVar(&roleDeleteFlags.server, "server", "", "Server ID (defaults to configured server)")
	roleDeleteCmd.Flags().StringVar(&roleDeleteFlags.role, "role", "", "Role ID to delete (required)")
	roleDeleteCmd.Flags().BoolVarP(&roleDeleteFlags.yes, "yes", "y", false, "Skip confirmation prompt")
}
