package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		{"Create Events", discordgo.PermissionCreateEvents},
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

// roleExport is one role in the export/import JSON.
type roleExport struct {
	Name        string   `json:"name"`
	Color       string   `json:"color,omitempty"`
	Hoist       bool     `json:"hoist,omitempty"`
	Mentionable bool     `json:"mentionable,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// rolesExportFile is the top-level JSON document for role export/import.
type rolesExportFile struct {
	Roles []roleExport `json:"roles"`
}

var (
	roleExportFlags = struct {
		server string
		file   string
	}{}
)

var roleExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export roles to JSON",
	Long: `Export the server's roles as JSON to a file (roles.json by default).
Managed (integration/bot) roles are skipped.

Examples:
  disc role export
  disc role export --file roles.json
  disc role export --server 123456789 --file roles.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(roleExportFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		roles, err := client.Session().GuildRoles(serverID)
		if err != nil {
			return fmt.Errorf("failed to list roles: %w", err)
		}

		file := rolesExportFile{}
		for _, r := range roles {
			if r.Managed {
				continue
			}
			re := roleExport{Name: r.Name, Hoist: r.Hoist, Mentionable: r.Mentionable}
			if r.Color != 0 {
				re.Color = fmt.Sprintf("%06X", r.Color)
			}
			for _, p := range permissionNames() {
				if r.Permissions&p.bit != 0 {
					re.Permissions = append(re.Permissions, p.name)
				}
			}
			file.Roles = append(file.Roles, re)
		}

		data, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode roles: %w", err)
		}

		out := roleExportFlags.file
		if out == "" {
			out = "roles.json"
		}
		if err := os.WriteFile(out, append(data, '\n'), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", out, err)
		}
		util.Green.Printf("Exported roles to %s\n", out)
		return nil
	},
}

var (
	roleImportFlags = struct {
		server        string
		file          string
		deleteMissing bool
		yes           bool
	}{}
)

var roleImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import roles from JSON",
	Long: `Apply a JSON role definition to the server, matching roles by name.
Reads roles.json by default.

Roles already present with the correct settings are left unchanged. Pass
--delete-missing to also delete roles on the server that are not in the file
(a non-reversible action). Managed roles and @everyone are never deleted. A
dry run is shown first; type "apply" to proceed.

Examples:
  disc role import
  disc role import --file roles.json
  disc role import --file roles.json --delete-missing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file := roleImportFlags.file
		if file == "" {
			file = "roles.json"
		}
		var f rolesExportFile
		if err := readJSONFile(file, &f); err != nil {
			return err
		}

		client, serverID, err := newClientAndServer(roleImportFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		roles, err := client.Session().GuildRoles(serverID)
		if err != nil {
			return fmt.Errorf("failed to list roles: %w", err)
		}

		changes, ops, err := planRoleImport(client.Session(), serverID, f, roles, roleImportFlags.deleteMissing)
		if err != nil {
			return err
		}
		abs, err := filepath.Abs(file)
		if err != nil {
			abs = file
		}
		if !applyPlan(abs, changes, roleImportFlags.yes) {
			util.Yellow.Println("Aborted.")
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

// permissionBitsByName builds a lookup of permission name to bit, so an
// exported display name can be turned back into a permission flag.
func permissionBitsByName() map[string]int64 {
	m := map[string]int64{}
	for _, p := range permissionNames() {
		m[p.name] = p.bit
	}
	return m
}

// computeRolePerms resolves an exported permission list into a flag bitmask,
// warning about any names that are not recognized.
func computeRolePerms(perms []string, permBits map[string]int64) (int64, error) {
	var bits int64
	for _, name := range perms {
		bit, ok := permBits[name]
		if !ok {
			return 0, fmt.Errorf("unknown permission '%s'", name)
		}
		bits |= bit
	}
	return bits, nil
}

// roleNeedsUpdate reports whether a live role differs from the desired state.
func roleNeedsUpdate(r *discordgo.Role, want roleExport, wantPerms int64) bool {
	wantColor := 0
	if want.Color != "" {
		wantColor, _ = parseHexColor(want.Color)
	}
	return r.Color != wantColor ||
		r.Hoist != want.Hoist ||
		r.Mentionable != want.Mentionable ||
		r.Permissions != wantPerms
}

// planRoleImport diffs the desired file against the live roles and returns the
// display plan plus the ordered operations to execute.
func planRoleImport(s *discordgo.Session, serverID string, file rolesExportFile, roles []*discordgo.Role, deleteMissing bool) ([]planAction, []func() error, error) {
	permBits := permissionBitsByName()

	desiredNames := map[string]bool{}
	for _, re := range file.Roles {
		if re.Name == "" {
			return nil, nil, fmt.Errorf("role entry is missing a name")
		}
		desiredNames[re.Name] = true
	}

	existingByName := map[string]*discordgo.Role{}
	for _, r := range roles {
		if r.Managed {
			continue
		}
		existingByName[r.Name] = r
	}

	var changes []planAction
	var ops []func() error

	for _, re := range file.Roles {
		re := re
		wantPerms, err := computeRolePerms(re.Permissions, permBits)
		if err != nil {
			return nil, nil, fmt.Errorf("role '%s': %w", re.Name, err)
		}
		wantColor := 0
		if re.Color != "" {
			c, err := parseHexColor(re.Color)
			if err != nil {
				return nil, nil, fmt.Errorf("role '%s': %w", re.Name, err)
			}
			wantColor = c
		}

		if er, exists := existingByName[re.Name]; exists {
			if roleNeedsUpdate(er, re, wantPerms) {
				changes = append(changes, planAction{"update", fmt.Sprintf("role '%s'", re.Name)})
				ops = append(ops, func() error {
					params := &discordgo.RoleParams{
						Name:        re.Name,
						Color:       &wantColor,
						Hoist:       &re.Hoist,
						Mentionable: &re.Mentionable,
						Permissions: &wantPerms,
					}
					return editRole(s, serverID, er.ID, params)
				})
			}
		} else {
			changes = append(changes, planAction{"create", fmt.Sprintf("role '%s'", re.Name)})
			ops = append(ops, func() error {
				params := &discordgo.RoleParams{
					Name:        re.Name,
					Color:       &wantColor,
					Hoist:       &re.Hoist,
					Mentionable: &re.Mentionable,
					Permissions: &wantPerms,
				}
				return createRole(s, serverID, params)
			})
		}
	}

	if deleteMissing {
		for name, r := range existingByName {
			if desiredNames[name] || name == "@everyone" {
				continue
			}
			r := r
			changes = append(changes, planAction{"delete", fmt.Sprintf("role '%s'", name)})
			ops = append(ops, func() error { return deleteRole(s, serverID, r.ID) })
		}
	}

	return changes, ops, nil
}

func createRole(s *discordgo.Session, serverID string, params *discordgo.RoleParams) error {
	role, err := s.GuildRoleCreate(serverID, params)
	if err != nil {
		return fmt.Errorf("failed to create role '%s': %w", params.Name, err)
	}
	util.Green.Printf("Created role %s (%s)\n", role.Name, role.ID)
	return nil
}

func editRole(s *discordgo.Session, serverID, roleID string, params *discordgo.RoleParams) error {
	role, err := s.GuildRoleEdit(serverID, roleID, params)
	if err != nil {
		return fmt.Errorf("failed to update role %s: %w", roleID, err)
	}
	util.Green.Printf("Updated role %s (%s)\n", role.Name, role.ID)
	return nil
}

func deleteRole(s *discordgo.Session, serverID, roleID string) error {
	if err := s.GuildRoleDelete(serverID, roleID); err != nil {
		return fmt.Errorf("failed to delete role %s: %w", roleID, err)
	}
	util.Green.Printf("Deleted role %s\n", roleID)
	return nil
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
	roleCmd.AddCommand(roleExportCmd)
	roleCmd.AddCommand(roleImportCmd)

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

	roleExportCmd.Flags().StringVar(&roleExportFlags.server, "server", "", "Server ID (defaults to configured server)")
	roleExportCmd.Flags().StringVar(&roleExportFlags.file, "file", "", "Output file (defaults to roles.json)")

	roleImportCmd.Flags().StringVar(&roleImportFlags.server, "server", "", "Server ID (defaults to configured server)")
	roleImportCmd.Flags().StringVar(&roleImportFlags.file, "file", "", "JSON file to import (defaults to roles.json)")
	roleImportCmd.Flags().BoolVar(&roleImportFlags.deleteMissing, "delete-missing", false, "Delete roles on the server not present in the file (non-reversible)")
	roleImportCmd.Flags().BoolVarP(&roleImportFlags.yes, "yes", "y", false, "Skip the 'apply' confirmation prompt")
}
