package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Manage Discord channels",
	Long:  "Commands for listing and managing Discord channels in a server.",
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List channels in a server",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(serverFlag)
		if err != nil {
			return err
		}
		defer client.Close()

		channels, err := client.Session().GuildChannels(serverID)
		if err != nil {
			return fmt.Errorf("failed to list channels: %w", err)
		}

		// Group channels by parent category.
		categories := map[string][]*discordgo.Channel{}
		var uncategorized []*discordgo.Channel
		for _, ch := range channels {
			if ch.Type == discordgo.ChannelTypeGuildCategory {
				continue
			}
			if ch.ParentID == "" || ch.ParentID == "0" {
				uncategorized = append(uncategorized, ch)
			} else {
				categories[ch.ParentID] = append(categories[ch.ParentID], ch)
			}
		}

		sortChannels := func(list []*discordgo.Channel) {
			sort.Slice(list, func(i, j int) bool { return list[i].Position < list[j].Position })
		}

		if len(uncategorized) > 0 {
			util.Bold.Println("(No Category)")
			sortChannels(uncategorized)
			for _, ch := range uncategorized {
				printChannel(ch)
			}
			fmt.Println()
		}

		// Order categories by position.
		catOrder := []*discordgo.Channel{}
		for _, ch := range channels {
			if ch.Type == discordgo.ChannelTypeGuildCategory {
				catOrder = append(catOrder, ch)
			}
		}
		sortChannels(catOrder)

		for _, cat := range catOrder {
			util.Bold.Println(cat.Name)
			children := categories[cat.ID]
			sortChannels(children)
			for _, ch := range children {
				printChannel(ch)
			}
			fmt.Println()
		}
		return nil
	},
}

func printChannel(ch *discordgo.Channel) {
	icon := ""
	switch ch.Type {
	case discordgo.ChannelTypeGuildVoice:
		icon = "🔊"
	default:
		icon = "💬"
	}
	fmt.Printf("  %s %s ", icon, ch.Name)
	util.Dim.Printf("(ID: %s)\n", ch.ID)
}

var (
	channelShowFlags = struct {
		server  string
		channel string
	}{}
)

var channelShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show channel details",
	Example: `  disc channel show --server 123456789 --channel 987654321`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if channelShowFlags.channel == "" {
			return fmt.Errorf("--channel is required")
		}
		client, serverID, err := newClientAndServer(channelShowFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		ch, err := client.Session().Channel(channelShowFlags.channel)
		if err != nil {
			return fmt.Errorf("failed to load channel: %w", err)
		}

		util.Bold.Printf("Channel: %s\n", ch.Name)
		util.Cyan.Printf("ID: %s\n", ch.ID)
		typeName, _ := channelTypeName(ch.Type)
		util.Cyan.Printf("Type: %s\n", typeName)
		util.Cyan.Printf("Position: %d\n", ch.Position)
		util.Cyan.Printf("NSFW: %t\n", ch.NSFW)
		util.Cyan.Printf("Topic: %s\n", ch.Topic)
		parent := "(none)"
		if ch.ParentID != "" && ch.ParentID != "0" {
			if p, err := client.Session().Channel(ch.ParentID); err == nil {
				parent = p.Name
			} else {
				parent = ch.ParentID
			}
		}
		util.Cyan.Printf("Category: %s\n", parent)

		roleID := map[string]string{}
		if roles, err := client.Session().GuildRoles(serverID); err == nil {
			for _, r := range roles {
				if r.ID == serverID {
					roleID[r.ID] = "@everyone"
				} else {
					roleID[r.ID] = r.Name
				}
			}
		}

		fmt.Println()
		util.Bold.Println("Permission overwrites:")
		if len(ch.PermissionOverwrites) == 0 {
			util.Dim.Println("  (none)")
		}
		for _, ow := range ch.PermissionOverwrites {
			subject := ow.ID
			if name, ok := roleID[ow.ID]; ok {
				subject = name
			}
			if ow.Type == discordgo.PermissionOverwriteTypeMember {
				subject = "member " + ow.ID
			}
			util.Cyan.Printf("  %s\n", subject)
			for _, p := range permNamesFromBits(ow.Allow) {
				util.Green.Printf("    + %s\n", p)
			}
			for _, p := range permNamesFromBits(ow.Deny) {
				util.Red.Printf("    - %s\n", p)
			}
		}
		return nil
	},
}

// permissionOverwriteExport is one role-based permission overwrite in the
// server config JSON. Roles are referenced by name; member overwrites are not
// managed.
type permissionOverwriteExport struct {
	Role  string   `json:"role"`            // role name, or "@everyone"
	Allow []string `json:"allow,omitempty"` // permission names
	Deny  []string `json:"deny,omitempty"`  // permission names
}

// channelExport is one channel in the server config JSON.
type channelExport struct {
	ID         string                      `json:"id,omitempty"`
	Name       string                      `json:"name"`
	Type       string                      `json:"type"`
	Topic      string                      `json:"topic,omitempty"`
	NSFW       bool                        `json:"nsfw,omitempty"`
	Parent     string                      `json:"parent,omitempty"` // parent category name
	Position   int                         `json:"position,omitempty"`
	Overwrites []permissionOverwriteExport `json:"overwrites,omitempty"`
}

// channelTypeName returns a canonical JSON name for a channel type. The bool
// is false for channel kinds that export/import does not manage (threads, etc).
func channelTypeName(t discordgo.ChannelType) (string, bool) {
	switch t {
	case discordgo.ChannelTypeGuildText:
		return "text", true
	case discordgo.ChannelTypeGuildVoice:
		return "voice", true
	case discordgo.ChannelTypeGuildCategory:
		return "category", true
	case discordgo.ChannelTypeGuildNews:
		return "news", true
	case discordgo.ChannelTypeGuildStageVoice:
		return "stage", true
	case discordgo.ChannelTypeGuildForum:
		return "forum", true
	default:
		return "", false
	}
}

// channelTypeFromName maps a JSON channel type back to a discordgo type.
func channelTypeFromName(s string) (discordgo.ChannelType, bool) {
	switch s {
	case "text":
		return discordgo.ChannelTypeGuildText, true
	case "voice":
		return discordgo.ChannelTypeGuildVoice, true
	case "category":
		return discordgo.ChannelTypeGuildCategory, true
	case "news":
		return discordgo.ChannelTypeGuildNews, true
	case "stage":
		return discordgo.ChannelTypeGuildStageVoice, true
	case "forum":
		return discordgo.ChannelTypeGuildForum, true
	default:
		return 0, false
	}
}

// permNamesFromBits returns the display names of all set permission bits.
func permNamesFromBits(bits int64) []string {
	var names []string
	for _, p := range permissionNames() {
		if bits&p.bit != 0 {
			names = append(names, p.name)
		}
	}
	return names
}

// exportChannels converts the server's live channels into config entries,
// skipping channel kinds that are not managed (threads, etc.). Each entry
// carries its server ID, its parent category name, and role-based permission
// overwrites. Member overwrites are not managed and are skipped.
func exportChannels(s *discordgo.Session, serverID string) ([]channelExport, error) {
	channels, err := s.GuildChannels(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	byID := map[string]*discordgo.Channel{}
	for _, ch := range channels {
		byID[ch.ID] = ch
	}

	roleNameByID := map[string]string{"@everyone": ""}
	if roles, err := s.GuildRoles(serverID); err == nil {
		for _, r := range roles {
			if r.ID == serverID {
				roleNameByID[r.ID] = "@everyone"
			} else {
				roleNameByID[r.ID] = r.Name
			}
		}
	}

	var out []channelExport
	for _, ch := range channels {
		name, ok := channelTypeName(ch.Type)
		if !ok {
			continue
		}
		ce := channelExport{ID: ch.ID, Name: ch.Name, Type: name, Topic: ch.Topic, NSFW: ch.NSFW, Position: ch.Position}
		if ch.ParentID != "" && ch.ParentID != "0" {
			if p := byID[ch.ParentID]; p != nil {
				ce.Parent = p.Name
			}
		}
		for _, ow := range ch.PermissionOverwrites {
			if ow.Type != discordgo.PermissionOverwriteTypeRole {
				continue
			}
			roleName, ok := roleNameByID[ow.ID]
			if !ok || roleName == "" {
				continue
			}
			ce.Overwrites = append(ce.Overwrites, permissionOverwriteExport{
				Role:  roleName,
				Allow: permNamesFromBits(ow.Allow),
				Deny:  permNamesFromBits(ow.Deny),
			})
		}
		out = append(out, ce)
	}
	return out, nil
}

// isSystemChannel reports whether a channel is reserved by the guild so it is
// never deleted by a push.
func isSystemChannel(guild *discordgo.Guild, ch *discordgo.Channel) bool {
	switch ch.ID {
	case guild.SystemChannelID, guild.RulesChannelID, guild.PublicUpdatesChannelID, guild.AfkChannelID:
		return true
	}
	return false
}

// planChannelImport diffs the desired channels against the live server state
// and returns the display plan plus the ordered operations to execute. Each
// desired channel is matched to a live channel by ID when it carries one,
// falling back to name for entries without an ID. roleIDByName maps role names
// to IDs (including "@everyone") and is updated as roles are created during the
// push so channel permission overwrites can resolve new roles.
func planChannelImport(s *discordgo.Session, guild *discordgo.Guild, desired []channelExport, roleIDByName map[string]string, deleteMissing bool) ([]planAction, []func() error, error) {
	live, err := s.GuildChannels(guild.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list channels: %w", err)
	}

	existingCatByID := map[string]*discordgo.Channel{}
	existingCatByName := map[string]*discordgo.Channel{}
	existingChByID := map[string]*discordgo.Channel{}
	existingChByName := map[string]*discordgo.Channel{}
	for _, ch := range live {
		if ch.Type == discordgo.ChannelTypeGuildCategory {
			existingCatByID[ch.ID] = ch
			existingCatByName[ch.Name] = ch
		} else {
			existingChByID[ch.ID] = ch
			existingChByName[ch.Name] = ch
		}
	}

	desiredCatNames := map[string]bool{}
	desiredChNames := map[string]bool{}
	desiredChIDs := map[string]bool{}
	var desiredCats, desiredChs []channelExport
	for _, ce := range desired {
		typ, ok := channelTypeFromName(ce.Type)
		if !ok {
			return nil, nil, fmt.Errorf("unsupported channel type '%s' for channel '%s'", ce.Type, ce.Name)
		}
		if ce.Name == "" {
			return nil, nil, fmt.Errorf("channel entry is missing a name")
		}
		if ce.ID != "" {
			desiredChIDs[ce.ID] = true
		}
		if typ == discordgo.ChannelTypeGuildCategory {
			desiredCatNames[ce.Name] = true
			desiredCats = append(desiredCats, ce)
		} else {
			desiredChNames[ce.Name] = true
			desiredChs = append(desiredChs, ce)
		}
	}

	var changes []planAction
	var ops []func() error

	// liveCatID tracks category name -> ID as categories are created so child
	// channels can resolve their parent during execution.
	liveCatID := map[string]string{}
	for name, cat := range existingCatByName {
		liveCatID[name] = cat.ID
	}

	// 1. Create / update categories.
	for _, dc := range desiredCats {
		dc := dc
		var ec *discordgo.Channel
		if dc.ID != "" {
			ec = existingCatByID[dc.ID]
		}
		if ec == nil {
			ec = existingCatByName[dc.Name]
		}
		if ec != nil {
			resolved := resolveOverwrites(ec.PermissionOverwrites, dc.Overwrites, roleIDByName, deleteMissing)
			if ec.Topic != dc.Topic || ec.NSFW != dc.NSFW || dc.Position != ec.Position || !overwritesEqual(ec.PermissionOverwrites, resolved) {
				ec := ec
				changes = append(changes, planAction{"update", fmt.Sprintf("category '%s'", dc.Name)})
				ops = append(ops, func() error {
					edit := &discordgo.ChannelEdit{Topic: dc.Topic, NSFW: &dc.NSFW, Position: &dc.Position}
					edit.PermissionOverwrites = resolveOverwrites(ec.PermissionOverwrites, dc.Overwrites, roleIDByName, deleteMissing)
					_, err := editChannel(s, ec.ID, edit)
					return err
				})
			}
		} else {
			changes = append(changes, planAction{"create", fmt.Sprintf("category '%s'", dc.Name)})
			ops = append(ops, func() error {
				data := discordgo.GuildChannelCreateData{Name: dc.Name, Type: discordgo.ChannelTypeGuildCategory, Topic: dc.Topic, NSFW: dc.NSFW, Position: dc.Position}
				data.PermissionOverwrites = resolveOverwrites(nil, dc.Overwrites, roleIDByName, deleteMissing)
				ch, err := createChannel(s, guild.ID, data)
				if err != nil {
					return err
				}
				liveCatID[dc.Name] = ch.ID
				return nil
			})
		}
	}

	// 2. Create / update channels.
	for _, dc := range desiredChs {
		dc := dc
		typ, _ := channelTypeFromName(dc.Type)
		parentID := resolveParentID(dc.Parent, liveCatID)
		var ec *discordgo.Channel
		if dc.ID != "" {
			ec = existingChByID[dc.ID]
		}
		if ec == nil {
			ec = existingChByName[dc.Name]
		}
		if ec != nil {
			resolved := resolveOverwrites(ec.PermissionOverwrites, dc.Overwrites, roleIDByName, deleteMissing)
			if channelNeedsUpdate(ec, typ, dc, parentID) || dc.Position != ec.Position || !overwritesEqual(ec.PermissionOverwrites, resolved) {
				ec := ec
				changes = append(changes, planAction{"update", fmt.Sprintf("channel '%s'", dc.Name)})
				ops = append(ops, func() error {
					edit := &discordgo.ChannelEdit{Name: dc.Name, Topic: dc.Topic, NSFW: &dc.NSFW, ParentID: parentID, Position: &dc.Position}
					edit.PermissionOverwrites = resolveOverwrites(ec.PermissionOverwrites, dc.Overwrites, roleIDByName, deleteMissing)
					_, err := editChannel(s, ec.ID, edit)
					return err
				})
			}
		} else {
			changes = append(changes, planAction{"create", fmt.Sprintf("channel '%s'", dc.Name)})
			ops = append(ops, func() error {
				data := discordgo.GuildChannelCreateData{Name: dc.Name, Type: typ, Topic: dc.Topic, NSFW: dc.NSFW, ParentID: parentID, Position: dc.Position}
				data.PermissionOverwrites = resolveOverwrites(nil, dc.Overwrites, roleIDByName, deleteMissing)
				_, err := createChannel(s, guild.ID, data)
				return err
			})
		}
	}

	if deleteMissing {
		// 3. Delete channels not in the file (children before categories).
		channelDeleteIDs := map[string]bool{}
		var deleteChIDs []string
		for name, ec := range existingChByName {
			if desiredChNames[name] || desiredChIDs[ec.ID] || isSystemChannel(guild, ec) {
				continue
			}
			channelDeleteIDs[ec.ID] = true
			deleteChIDs = append(deleteChIDs, ec.ID)
			changes = append(changes, planAction{"delete", fmt.Sprintf("channel '%s'", name)})
		}
		for _, id := range deleteChIDs {
			id := id
			ops = append(ops, func() error { return deleteChannel(s, id) })
		}

		// 4. Delete categories not in the file, only if none of their children
		// are being kept (deleting a category cascades to its children).
		for name, ec := range existingCatByName {
			if desiredCatNames[name] || desiredChIDs[ec.ID] {
				continue
			}
			hasKeptChild := false
			for _, child := range live {
				if child.ParentID == ec.ID && !channelDeleteIDs[child.ID] {
					hasKeptChild = true
					break
				}
			}
			if hasKeptChild {
				changes = append(changes, planAction{"update", fmt.Sprintf("category '%s' (kept: has retained children)", name)})
				continue
			}
			ec := ec
			changes = append(changes, planAction{"delete", fmt.Sprintf("category '%s'", name)})
			ops = append(ops, func() error { return deleteChannel(s, ec.ID) })
		}
	}

	return changes, ops, nil
}

func channelNeedsUpdate(ec *discordgo.Channel, wantType discordgo.ChannelType, want channelExport, wantParentID string) bool {
	existingType, _ := channelTypeName(ec.Type)
	existingParent := ""
	if ec.ParentID != "" && ec.ParentID != "0" {
		existingParent = ec.ParentID
	}
	return existingType != want.Type ||
		ec.Topic != want.Topic ||
		ec.NSFW != want.NSFW ||
		existingParent != wantParentID
}

// resolveOverwrites computes the final permission overwrites for a channel
// from the desired config overwrites and the live overwrites. Member
// overwrites are always preserved. Live role overwrites absent from the config
// are kept unless deleteMissing. roleIDByName maps role names to IDs
// (including "@everyone"); roles that cannot be resolved yet are skipped.
func resolveOverwrites(live []*discordgo.PermissionOverwrite, desired []permissionOverwriteExport, roleIDByName map[string]string, deleteMissing bool) []*discordgo.PermissionOverwrite {
	desiredByID := map[string]*discordgo.PermissionOverwrite{}
	var ordered []*discordgo.PermissionOverwrite
	for _, d := range desired {
		id, ok := roleIDByName[d.Role]
		if !ok {
			continue
		}
		ow := &discordgo.PermissionOverwrite{ID: id, Type: discordgo.PermissionOverwriteTypeRole, Allow: permBitsFor(d.Allow), Deny: permBitsFor(d.Deny)}
		if _, exists := desiredByID[id]; !exists {
			ordered = append(ordered, ow)
		}
		desiredByID[id] = ow
	}

	var out []*discordgo.PermissionOverwrite
	for _, ow := range live {
		if ow.Type == discordgo.PermissionOverwriteTypeMember {
			out = append(out, ow)
			continue
		}
		if desiredByID[ow.ID] == nil && !deleteMissing {
			out = append(out, ow)
		}
	}
	return append(out, ordered...)
}

// overwritesEqual reports whether two permission overwrite slices match
// (order-insensitive), ignoring overwrites for roles that cannot be resolved.
func overwritesEqual(a, b []*discordgo.PermissionOverwrite) bool {
	key := func(ow *discordgo.PermissionOverwrite) string {
		return fmt.Sprintf("%s:%d:%d:%d", ow.ID, ow.Type, ow.Allow, ow.Deny)
	}
	count := func(list []*discordgo.PermissionOverwrite) map[string]int {
		m := map[string]int{}
		for _, ow := range list {
			m[key(ow)]++
		}
		return m
	}
	ma, mb := count(a), count(b)
	if len(ma) != len(mb) {
		return false
	}
	for k, v := range ma {
		if mb[k] != v {
			return false
		}
	}
	return true
}

func resolveParentID(parentName string, liveCatID map[string]string) string {
	if parentName == "" {
		return ""
	}
	return liveCatID[parentName]
}

func createChannel(s *discordgo.Session, guildID string, data discordgo.GuildChannelCreateData) (*discordgo.Channel, error) {
	ch, err := s.GuildChannelCreateComplex(guildID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel '%s': %w", data.Name, err)
	}
	util.Green.Printf("Created channel %s (%s)\n", ch.Name, ch.ID)
	return ch, nil
}

func editChannel(s *discordgo.Session, id string, edit *discordgo.ChannelEdit) (*discordgo.Channel, error) {
	ch, err := s.ChannelEdit(id, edit)
	if err != nil {
		return nil, fmt.Errorf("failed to update channel %s: %w", id, err)
	}
	util.Green.Printf("Updated channel %s (%s)\n", ch.Name, ch.ID)
	return ch, nil
}

func deleteChannel(s *discordgo.Session, id string) error {
	if _, err := s.ChannelDelete(id); err != nil {
		return fmt.Errorf("failed to delete channel %s: %w", id, err)
	}
	util.Green.Printf("Deleted channel %s\n", id)
	return nil
}

// permTarget holds the allow/deny sets being applied to a single role, plus
// which side the caller actually provided (so the un-provided side can be
// preserved on an existing overwrite).
type permTarget struct {
	allow        []string
	deny         []string
	allowChanged bool
	denyChanged  bool
}

// collectPermTargets merges repeated --allow/--deny values (each "Role:Perm,Perm")
// into a per-role allow/deny map.
func collectPermTargets(allows, denies []string) (map[string]*permTarget, error) {
	targets := map[string]*permTarget{}
	add := func(entries []string, isAllow bool) error {
		for _, e := range entries {
			role, permStr, ok := strings.Cut(e, ":")
			if !ok {
				return fmt.Errorf("expected 'Role:Perm,Perm' but got '%s'", e)
			}
			role = strings.TrimSpace(role)
			if role == "" {
				return fmt.Errorf("missing role in '%s'", e)
			}
			perms, err := splitPermissions(permStr)
			if err != nil {
				return fmt.Errorf("role '%s': %w", role, err)
			}
			t := targets[role]
			if t == nil {
				t = &permTarget{}
				targets[role] = t
			}
			if isAllow {
				t.allow = append(t.allow, perms...)
				t.allowChanged = true
			} else {
				t.deny = append(t.deny, perms...)
				t.denyChanged = true
			}
		}
		return nil
	}
	if err := add(allows, true); err != nil {
		return nil, err
	}
	if err := add(denies, false); err != nil {
		return nil, err
	}
	return targets, nil
}

// permBitsFor converts permission names to a bitmask.
func permBitsFor(names []string) int64 {
	permBits := permissionBitsByName()
	var bits int64
	for _, n := range names {
		if b, ok := permBits[n]; ok {
			bits |= b
		}
	}
	return bits
}

// roleIDByName maps role names to IDs, with "@everyone" mapping to the guild ID.
func roleIDByName(s *discordgo.Session, serverID string) (map[string]string, error) {
	roles, err := s.GuildRoles(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	m := map[string]string{}
	for _, r := range roles {
		if r.ID == serverID {
			m["@everyone"] = r.ID
		} else {
			m[r.Name] = r.ID
		}
	}
	return m, nil
}

// overwritesForTargets resolves targeted roles to IDs and returns the
// permission overwrites for them.
var serverFlag string

func init() {
	channelCmd.AddCommand(channelListCmd)
	channelCmd.AddCommand(channelShowCmd)

	channelListCmd.Flags().StringVar(&serverFlag, "server", "", "Server ID (defaults to configured server)")

	channelShowCmd.Flags().StringVar(&channelShowFlags.server, "server", "", "Server ID (defaults to configured server)")
	channelShowCmd.Flags().StringVar(&channelShowFlags.channel, "channel", "", "Channel ID to show details for (required)")
}
