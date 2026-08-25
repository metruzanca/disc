package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

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
	Use:     "show",
	Short:   "Show role details",
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

// permissionNameStrings returns just the names from permissionNames().
func permissionNameStrings() []string {
	names := permissionNames()
	out := make([]string, 0, len(names))
	for _, p := range names {
		out = append(out, p.name)
	}
	return out
}

// roleExport is one role in the server config JSON.
type roleExport struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Color       string   `json:"color,omitempty"`
	Hoist       bool     `json:"hoist,omitempty"`
	Mentionable bool     `json:"mentionable,omitempty"`
	Position    int      `json:"position,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// exportRoles converts the server's live roles into config entries, skipping
// managed (integration/bot) roles. Each entry carries its server ID.
func exportRoles(s *discordgo.Session, serverID string) ([]roleExport, error) {
	roles, err := s.GuildRoles(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	var out []roleExport
	for _, r := range roles {
		if r.Managed {
			continue
		}
		re := roleExport{ID: r.ID, Name: r.Name, Hoist: r.Hoist, Mentionable: r.Mentionable, Position: r.Position}
		if r.Color != 0 {
			re.Color = fmt.Sprintf("%06X", r.Color)
		}
		for _, p := range permissionNames() {
			if r.Permissions&p.bit != 0 {
				re.Permissions = append(re.Permissions, p.name)
			}
		}
		out = append(out, re)
	}
	return out, nil
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

// planRoleImport diffs the desired roles against the live roles and returns
// the display plan plus the ordered operations to execute. Each desired role is
// matched to a live role by ID when it carries one, falling back to name for
// entries without an ID (e.g. freshly added config entries). roleIDByName maps
// role names to IDs and is updated as roles are created. Roles are reordered to
// match their configured positions when the order differs.
func planRoleImport(s *discordgo.Session, serverID string, desired []roleExport, roles []*discordgo.Role, roleIDByName map[string]string, deleteMissing bool) ([]planAction, []func() error, error) {
	permBits := permissionBitsByName()

	desiredIDs := map[string]bool{}
	desiredNames := map[string]bool{}
	for _, re := range desired {
		if re.Name == "" {
			return nil, nil, fmt.Errorf("role entry is missing a name")
		}
		if re.ID != "" {
			desiredIDs[re.ID] = true
		}
		desiredNames[re.Name] = true
	}

	existingByID := map[string]*discordgo.Role{}
	existingByName := map[string]*discordgo.Role{}
	for _, r := range roles {
		if r.Managed {
			continue
		}
		existingByID[r.ID] = r
		existingByName[r.Name] = r
	}

	var changes []planAction
	var ops []func() error

	for _, re := range desired {
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

		// Resolve the matching live role: prefer ID, then name.
		var er *discordgo.Role
		if re.ID != "" {
			er = existingByID[re.ID]
		}
		if er == nil {
			er = existingByName[re.Name]
		}

		params := &discordgo.RoleParams{
			Name:        re.Name,
			Color:       &wantColor,
			Hoist:       &re.Hoist,
			Mentionable: &re.Mentionable,
			Permissions: &wantPerms,
		}
		if er != nil {
			if roleNeedsUpdate(er, re, wantPerms) {
				er := er
				changes = append(changes, planAction{"update", fmt.Sprintf("role '%s'", re.Name)})
				ops = append(ops, func() error {
					return editRole(s, serverID, er.ID, params)
				})
			}
		} else {
			changes = append(changes, planAction{"create", fmt.Sprintf("role '%s'", re.Name)})
			ops = append(ops, func() error {
				return createRole(s, serverID, params, roleIDByName)
			})
		}
	}

	if deleteMissing {
		for name, r := range existingByName {
			if desiredNames[name] || desiredIDs[r.ID] || name == "@everyone" {
				continue
			}
			r := r
			changes = append(changes, planAction{"delete", fmt.Sprintf("role '%s'", name)})
			ops = append(ops, func() error { return deleteRole(s, serverID, r.ID) })
		}
	}

	if reorder, op := roleReorderOp(s, serverID, desired, roles); reorder {
		changes = append(changes, planAction{"update", "role order"})
		ops = append(ops, op)
	}

	return changes, ops, nil
}

// roleReorderOp detects whether the configured role order differs from the live
// order and, if so, returns an operation that reorders the roles. It only
// considers roles present in the config (by ID). Managed roles and roles not in
// the config are kept in their current slots.
func roleReorderOp(s *discordgo.Session, serverID string, desired []roleExport, roles []*discordgo.Role) (bool, func() error) {
	pos := map[string]int{}
	for _, re := range desired {
		if re.ID != "" {
			pos[re.ID] = re.Position
		}
	}
	if len(pos) == 0 {
		return false, nil
	}

	// Desired order: highest position first (matches Discord hierarchy).
	ids := make([]string, 0, len(pos))
	for id := range pos {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return pos[ids[i]] > pos[ids[j]] })

	liveRank := map[string]int{}
	for i, r := range roles {
		liveRank[r.ID] = i
	}
	changed := false
	for i := 1; i < len(ids); i++ {
		if liveRank[ids[i-1]] > liveRank[ids[i]] {
			changed = true
			break
		}
	}
	if !changed {
		return false, nil
	}

	inDesired := map[string]bool{}
	for _, id := range ids {
		inDesired[id] = true
	}
	var target []*discordgo.Role
	var deferred []*discordgo.Role
	for _, r := range roles {
		if inDesired[r.ID] {
			deferred = append(deferred, r)
		} else {
			target = append(target, r)
		}
	}
	sort.Slice(deferred, func(i, j int) bool { return pos[deferred[i].ID] > pos[deferred[j].ID] })
	target = append(target, deferred...)

	return true, func() error {
		if _, err := s.GuildRoleReorder(serverID, target); err != nil {
			return fmt.Errorf("failed to reorder roles: %w", err)
		}
		util.Green.Println("Reordered roles")
		return nil
	}
}

func createRole(s *discordgo.Session, serverID string, params *discordgo.RoleParams, roleIDByName map[string]string) error {
	role, err := s.GuildRoleCreate(serverID, params)
	if err != nil {
		return fmt.Errorf("failed to create role '%s': %w", params.Name, err)
	}
	roleIDByName[params.Name] = role.ID
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

func loadRoleConfigFileForEdit(fileFlag string, create bool) (*serverConfigFile, string, error) {
	path, err := resolveConfigFile(fileFlag, "")
	if err != nil {
		return nil, "", err
	}
	var cfg serverConfigFile
	if err := readJSONFile(path, &cfg); err != nil {
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

var rolePermCmd = &cobra.Command{
	Use:   "perm",
	Short: "List permission names",
	Long:  "Utilities for working with Discord permission names.",
}

var rolePermListCmd = &cobra.Command{
	Use:   "list",
	Short: "List known permission names",
	Long: `List every recognized permission name, ordered from most to least
privileged. Use these exact names with disc role show, or with
disc config role add/update --permissions.

Example:
  disc role perm list`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, p := range permissionNames() {
			util.Bold.Printf("  %s\n", p.name)
		}
		return nil
	},
}

var (
	roleAddFlags = struct {
		file        string
		name        string
		color       string
		hoist       bool
		mentionable bool
		permissions string
	}{}
	roleUpFlags = struct {
		file        string
		role        string
		name        string
		color       string
		hoist       bool
		mentionable bool
		permissions string
		position    int
	}{}
	roleDelFlags = struct {
		file string
		role string
	}{}
)

var roleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a role to the config",
	Long: `Add a role to the config file. The role is not created on the server
until 'disc server push' runs. If a role with the same name already exists in
the config, an error is returned.

Examples:
  disc role add --name Moderator
  disc role add --name VIP --color FF0000 --hoist
  disc role add --name Helper --permissions "Send Messages,Add Reactions"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := loadConfigFileForEdit(roleAddFlags.file, "", true)
		if err != nil {
			return err
		}
		if roleAddFlags.name == "" {
			return fmt.Errorf("--name is required")
		}
		for _, r := range cfg.Roles {
			if r.Name == roleAddFlags.name {
				return fmt.Errorf("role '%s' already exists in config; use 'disc role update' instead", r.Name)
			}
		}

		re := roleExport{Name: roleAddFlags.name, Hoist: roleAddFlags.hoist, Mentionable: roleAddFlags.mentionable}
		if roleAddFlags.color != "" {
			if _, err := parseHexColor(roleAddFlags.color); err != nil {
				return err
			}
			re.Color = strings.ToUpper(roleAddFlags.color)
		}
		if cmd.Flags().Changed("permissions") {
			perms, err := splitPermissions(roleAddFlags.permissions)
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

var roleUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a role in the config",
	Long: `Update a role in the config file by its ID. The change is applied to
the server on the next 'disc server push'.

Examples:
  disc role update --role 987654321 --name "New Name"
  disc role update --role 987654321 --color 00FF00
  disc role update --role 987654321 --permissions "Kick Members,Ban Members"
  disc role update --role 987654321 --position 3`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := loadConfigFileForEdit(roleUpFlags.file, "", false)
		if err != nil {
			return err
		}
		i := findRoleInConfig(cfg, roleUpFlags.role)
		if i == -1 {
			return fmt.Errorf("role with ID '%s' not found in config", roleUpFlags.role)
		}

		re := &cfg.Roles[i]
		if cmd.Flags().Changed("name") {
			re.Name = roleUpFlags.name
		}
		if cmd.Flags().Changed("color") {
			if roleUpFlags.color == "" {
				re.Color = ""
			} else {
				if _, err := parseHexColor(roleUpFlags.color); err != nil {
					return err
				}
				re.Color = strings.ToUpper(roleUpFlags.color)
			}
		}
		if cmd.Flags().Changed("hoist") {
			re.Hoist = roleUpFlags.hoist
		}
		if cmd.Flags().Changed("mentionable") {
			re.Mentionable = roleUpFlags.mentionable
		}
		if cmd.Flags().Changed("permissions") {
			perms, err := splitPermissions(roleUpFlags.permissions)
			if err != nil {
				return err
			}
			re.Permissions = perms
		}
		if cmd.Flags().Changed("position") {
			re.Position = roleUpFlags.position
		}

		if err := writeConfigFile(cfg, path); err != nil {
			return err
		}
		util.Green.Printf("Updated role '%s' in %s\n", re.Name, path)
		return nil
	},
}

var roleDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a role from the config",
	Long: `Remove a role from the config file by its ID. The role is removed
from the server on the next 'disc server push'.

Example:
  disc role delete --role 987654321`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := loadConfigFileForEdit(roleDelFlags.file, "", false)
		if err != nil {
			return err
		}
		i := findRoleInConfig(cfg, roleDelFlags.role)
		if i == -1 {
			return fmt.Errorf("role with ID '%s' not found in config", roleDelFlags.role)
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
	roleCmd.AddCommand(roleListCmd)
	roleCmd.AddCommand(roleShowCmd)
	roleCmd.AddCommand(rolePermCmd)
	rolePermCmd.AddCommand(rolePermListCmd)
	roleCmd.AddCommand(roleAddCmd)
	roleCmd.AddCommand(roleUpdateCmd)
	roleCmd.AddCommand(roleDeleteCmd)

	roleListCmd.Flags().StringVar(&roleListFlags.server, "server", "", "Server ID (defaults to active server)")

	roleShowCmd.Flags().StringVar(&roleShowFlags.server, "server", "", "Server ID (defaults to active server)")
	roleShowCmd.Flags().StringVar(&roleShowFlags.role, "role", "", "Role ID to show details for (required)")

	roleAddCmd.Flags().StringVar(&roleAddFlags.file, "file", "", "Config file (defaults to resolved server.json)")
	roleAddCmd.Flags().StringVar(&roleAddFlags.name, "name", "", "Role name (required)")
	roleAddCmd.Flags().StringVar(&roleAddFlags.color, "color", "", "Role color in hex (e.g., FF0000 for red)")
	roleAddCmd.Flags().BoolVar(&roleAddFlags.hoist, "hoist", false, "Display role separately in member list")
	roleAddCmd.Flags().BoolVar(&roleAddFlags.mentionable, "mentionable", false, "Allow anyone to mention this role")
	roleAddCmd.Flags().StringVar(&roleAddFlags.permissions, "permissions", "", "Comma-separated permission names; see 'disc role perm list'")

	roleUpdateCmd.Flags().StringVar(&roleUpFlags.file, "file", "", "Config file (defaults to resolved server.json)")
	roleUpdateCmd.Flags().StringVar(&roleUpFlags.role, "role", "", "Role ID to update (required)")
	roleUpdateCmd.Flags().StringVar(&roleUpFlags.name, "name", "", "New role name")
	roleUpdateCmd.Flags().StringVar(&roleUpFlags.color, "color", "", "Role color in hex (e.g., FF0000 for red)")
	roleUpdateCmd.Flags().BoolVar(&roleUpFlags.hoist, "hoist", false, "Display role separately in member list")
	roleUpdateCmd.Flags().BoolVar(&roleUpFlags.mentionable, "mentionable", false, "Allow anyone to mention this role")
	roleUpdateCmd.Flags().StringVar(&roleUpFlags.permissions, "permissions", "", "Comma-separated permission names; see 'disc role perm list'")
	roleUpdateCmd.Flags().IntVar(&roleUpFlags.position, "position", 0, "Role position (hierarchy rank; higher is above)")

	roleDeleteCmd.Flags().StringVar(&roleDelFlags.file, "file", "", "Config file (defaults to resolved server.json)")
	roleDeleteCmd.Flags().StringVar(&roleDelFlags.role, "role", "", "Role ID to delete (required)")
}
