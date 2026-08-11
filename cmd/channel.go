package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	channelAddFlags = struct {
		server   string
		name     string
		typ      string
		category string
		yes      bool
	}{}
)

var channelAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new channel",
	Long: `Create a new channel in a Discord server.

Examples:
  disc channel add --server 123456789 --name general
  disc channel add --server 123456789 --name voice-chat --type voice
  disc channel add --server 123456789 --name help --category 987654321`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(channelAddFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()
		if channelAddFlags.name == "" {
			return fmt.Errorf("--name is required")
		}

		chType := discordgo.ChannelTypeGuildText
		if channelAddFlags.typ == "voice" {
			chType = discordgo.ChannelTypeGuildVoice
		}

		summary := fmt.Sprintf("Create %s channel '%s' in server %s?", channelAddFlags.typ, channelAddFlags.name, serverID)
		if !util.Confirm(summary, channelAddFlags.yes) {
			util.Yellow.Println("Aborted.")
			return nil
		}

		params := discordgo.GuildChannelCreateData{
			Name:     channelAddFlags.name,
			Type:     chType,
			ParentID: channelAddFlags.category,
		}
		ch, err := client.Session().GuildChannelCreateComplex(serverID, params)
		if err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}
		util.Green.Printf("Created channel %s (%s)\n", ch.Name, ch.ID)
		return nil
	},
}

var (
	channelUpdateFlags = struct {
		channel  string
		name     string
		topic    string
		category string
		nsfw     bool
		yes      bool
	}{}
)

var channelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a channel",
	Long: `Update an existing channel's properties.

Examples:
  disc channel update --channel 123456789 --name new-name
  disc channel update --channel 123456789 --topic "Welcome to the channel"
  disc channel update --channel 123456789 --category 987654321`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if channelUpdateFlags.channel == "" {
			return fmt.Errorf("--channel is required")
		}
		if channelUpdateFlags.name == "" && channelUpdateFlags.topic == "" && channelUpdateFlags.category == "" && !cmd.Flags().Changed("nsfw") {
			return fmt.Errorf("nothing to update; provide --name, --topic, --category, or --nsfw")
		}

		summary := fmt.Sprintf("Update channel %s?", channelUpdateFlags.channel)
		if !util.Confirm(summary, channelUpdateFlags.yes) {
			util.Yellow.Println("Aborted.")
			return nil
		}

		client, err := newClient()
		if err != nil {
			return err
		}
		defer client.Close()

		edits := &discordgo.ChannelEdit{}
		if channelUpdateFlags.name != "" {
			edits.Name = channelUpdateFlags.name
		}
		if channelUpdateFlags.topic != "" {
			edits.Topic = channelUpdateFlags.topic
		}
		if channelUpdateFlags.category != "" {
			edits.ParentID = channelUpdateFlags.category
		}
		if cmd.Flags().Changed("nsfw") {
			edits.NSFW = &channelUpdateFlags.nsfw
		}

		ch, err := client.Session().ChannelEdit(channelUpdateFlags.channel, edits)
		if err != nil {
			return fmt.Errorf("failed to update channel: %w", err)
		}
		util.Green.Printf("Updated channel %s (%s)\n", ch.Name, ch.ID)
		return nil
	},
}

var (
	channelDeleteFlags = struct {
		channel string
		yes     bool
	}{}
)

var channelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if channelDeleteFlags.channel == "" {
			return fmt.Errorf("--channel is required")
		}
		if !util.Confirm(fmt.Sprintf("Delete channel %s?", channelDeleteFlags.channel), channelDeleteFlags.yes) {
			util.Yellow.Println("Aborted.")
			return nil
		}

		client, err := newClient()
		if err != nil {
			return err
		}
		defer client.Close()

		if _, err := client.Session().ChannelDelete(channelDeleteFlags.channel); err != nil {
			return fmt.Errorf("failed to delete channel: %w", err)
		}
		util.Green.Printf("Deleted channel %s\n", channelDeleteFlags.channel)
		return nil
	},
}

var (
	channelMoveFlags = struct {
		channel  string
		position int
		category string
		server   string
		yes      bool
	}{}
)

var channelMoveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a channel to a different position or category",
	Long: `Move a channel to a different position or category.

Examples:
  disc channel move --server 123456789 --channel 111111111 --position 0
  disc channel move --server 123456789 --channel 111111111 --category 222222222
  disc channel move --server 123456789 --channel 111111111 --category 222222222 --position 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if channelMoveFlags.channel == "" {
			return fmt.Errorf("--channel is required")
		}

		summary := fmt.Sprintf("Move channel %s?", channelMoveFlags.channel)
		if !util.Confirm(summary, channelMoveFlags.yes) {
			util.Yellow.Println("Aborted.")
			return nil
		}

		client, err := newClient()
		if err != nil {
			return err
		}
		defer client.Close()

		ch, err := client.Session().Channel(channelMoveFlags.channel)
		if err != nil {
			return fmt.Errorf("failed to load channel: %w", err)
		}

		// ParentID "" means no category.
		parentID := ch.ParentID
		if channelMoveFlags.category != "" {
			if channelMoveFlags.category == "none" {
				parentID = ""
			} else {
				parentID = channelMoveFlags.category
			}
		}

		position := ch.Position
		if channelMoveFlags.position >= 0 {
			position = channelMoveFlags.position
		}

		if _, err := client.Session().ChannelEditComplex(channelMoveFlags.channel, &discordgo.ChannelEdit{ParentID: parentID, Position: &position}); err != nil {
			return fmt.Errorf("failed to move channel: %w", err)
		}
		util.Green.Printf("Moved channel %s\n", channelMoveFlags.channel)
		return nil
	},
}

// channelExport is one channel in the export/import JSON.
type channelExport struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Topic    string `json:"topic,omitempty"`
	NSFW     bool   `json:"nsfw,omitempty"`
	Parent   string `json:"parent,omitempty"` // parent category name
	Position int    `json:"position,omitempty"`
}

// channelsExportFile is the top-level JSON document for channel export/import.
type channelsExportFile struct {
	Channels []channelExport `json:"channels"`
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

var (
	channelExportFlags = struct {
		server string
		file   string
	}{}
)

var channelExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export channels to JSON",
	Long: `Export the server's channels as JSON to a file (channels.json by
default).

Examples:
  disc channel export
  disc channel export --file channels.json
  disc channel export --server 123456789 --file channels.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(channelExportFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		channels, err := client.Session().GuildChannels(serverID)
		if err != nil {
			return fmt.Errorf("failed to list channels: %w", err)
		}

		byID := map[string]*discordgo.Channel{}
		for _, ch := range channels {
			byID[ch.ID] = ch
		}

		file := channelsExportFile{}
		for _, ch := range channels {
			name, ok := channelTypeName(ch.Type)
			if !ok {
				continue
			}
			ce := channelExport{Name: ch.Name, Type: name, Topic: ch.Topic, NSFW: ch.NSFW, Position: ch.Position}
			if ch.ParentID != "" && ch.ParentID != "0" {
				if p := byID[ch.ParentID]; p != nil {
					ce.Parent = p.Name
				}
			}
			file.Channels = append(file.Channels, ce)
		}

		data, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode channels: %w", err)
		}

		out := channelExportFlags.file
		if out == "" {
			out = "channels.json"
		}
		if err := os.WriteFile(out, append(data, '\n'), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", out, err)
		}
		util.Green.Printf("Exported channels to %s\n", out)
		return nil
	},
}

var (
	channelImportFlags = struct {
		server        string
		file          string
		deleteMissing bool
		yes           bool
	}{}
)

var channelImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import channels from JSON",
	Long: `Apply a JSON channel definition to the server, matching channels by name.
Reads channels.json by default.

Channels already present with the correct settings are left unchanged. Pass
--delete-missing to also delete channels on the server that are not in the
file (a non-reversible action). A dry run is shown first; type "apply" to
proceed.

Examples:
  disc channel import
  disc channel import --file channels.json
  disc channel import --file channels.json --delete-missing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file := channelImportFlags.file
		if file == "" {
			file = "channels.json"
		}
		var f channelsExportFile
		if err := readJSONFile(file, &f); err != nil {
			return err
		}

		client, serverID, err := newClientAndServer(channelImportFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		guild, err := client.Session().Guild(serverID)
		if err != nil {
			return fmt.Errorf("failed to load server: %w", err)
		}

		changes, ops, err := planChannelImport(client.Session(), guild, f, channelImportFlags.deleteMissing)
		if err != nil {
			return err
		}
		abs, err := filepath.Abs(file)
		if err != nil {
			abs = file
		}
		if !applyPlan(abs, changes, channelImportFlags.yes) {
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

// isSystemChannel reports whether a channel is reserved by the guild so it is
// never deleted by an import.
func isSystemChannel(guild *discordgo.Guild, ch *discordgo.Channel) bool {
	switch ch.ID {
	case guild.SystemChannelID, guild.RulesChannelID, guild.PublicUpdatesChannelID, guild.AfkChannelID:
		return true
	}
	return false
}

// planChannelImport diffs the desired file against the live server state and
// returns the display plan plus the ordered operations to execute.
func planChannelImport(s *discordgo.Session, guild *discordgo.Guild, file channelsExportFile, deleteMissing bool) ([]planAction, []func() error, error) {
	live, err := s.GuildChannels(guild.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list channels: %w", err)
	}

	existingCatByName := map[string]*discordgo.Channel{}
	existingChByName := map[string]*discordgo.Channel{}
	for _, ch := range live {
		if ch.Type == discordgo.ChannelTypeGuildCategory {
			existingCatByName[ch.Name] = ch
		} else {
			existingChByName[ch.Name] = ch
		}
	}

	desiredCatNames := map[string]bool{}
	desiredChNames := map[string]bool{}
	var desiredCats, desiredChs []channelExport
	for _, ce := range file.Channels {
		typ, ok := channelTypeFromName(ce.Type)
		if !ok {
			return nil, nil, fmt.Errorf("unsupported channel type '%s' for channel '%s'", ce.Type, ce.Name)
		}
		if ce.Name == "" {
			return nil, nil, fmt.Errorf("channel entry is missing a name")
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
		if ec, exists := existingCatByName[dc.Name]; exists {
			if ec.Topic != dc.Topic || ec.NSFW != dc.NSFW {
				changes = append(changes, planAction{"update", fmt.Sprintf("category '%s'", dc.Name)})
				ops = append(ops, func() error {
					_, err := editChannel(s, ec.ID, &discordgo.ChannelEdit{Topic: dc.Topic, NSFW: &dc.NSFW})
					return err
				})
			}
		} else {
			changes = append(changes, planAction{"create", fmt.Sprintf("category '%s'", dc.Name)})
			ops = append(ops, func() error {
				ch, err := createChannel(s, guild.ID, discordgo.GuildChannelCreateData{Name: dc.Name, Type: discordgo.ChannelTypeGuildCategory, Topic: dc.Topic, NSFW: dc.NSFW})
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
		if ec, exists := existingChByName[dc.Name]; exists {
			if channelNeedsUpdate(ec, typ, dc, parentID) {
				changes = append(changes, planAction{"update", fmt.Sprintf("channel '%s'", dc.Name)})
				ops = append(ops, func() error {
					edit := &discordgo.ChannelEdit{Name: dc.Name, Topic: dc.Topic, NSFW: &dc.NSFW, ParentID: parentID}
					_, err := editChannel(s, ec.ID, edit)
					return err
				})
			}
		} else {
			changes = append(changes, planAction{"create", fmt.Sprintf("channel '%s'", dc.Name)})
			ops = append(ops, func() error {
				_, err := createChannel(s, guild.ID, discordgo.GuildChannelCreateData{Name: dc.Name, Type: typ, Topic: dc.Topic, NSFW: dc.NSFW, ParentID: parentID})
				return err
			})
		}
	}

	if deleteMissing {
		// 3. Delete channels not in the file (children before categories).
		channelDeleteIDs := map[string]bool{}
		var deleteChIDs []string
		for name, ec := range existingChByName {
			if desiredChNames[name] || isSystemChannel(guild, ec) {
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
			if desiredCatNames[name] {
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

var serverFlag string

func init() {
	channelCmd.AddCommand(channelListCmd)
	channelCmd.AddCommand(channelAddCmd)
	channelCmd.AddCommand(channelUpdateCmd)
	channelCmd.AddCommand(channelDeleteCmd)
	channelCmd.AddCommand(channelMoveCmd)
	channelCmd.AddCommand(channelExportCmd)
	channelCmd.AddCommand(channelImportCmd)

	channelListCmd.Flags().StringVar(&serverFlag, "server", "", "Server ID (defaults to configured server)")

	channelAddCmd.Flags().StringVar(&channelAddFlags.server, "server", "", "Server ID (defaults to configured server)")
	channelAddCmd.Flags().StringVar(&channelAddFlags.name, "name", "", "Channel name (required)")
	channelAddCmd.Flags().StringVar(&channelAddFlags.typ, "type", "text", "Channel type: text or voice")
	channelAddCmd.Flags().StringVar(&channelAddFlags.category, "category", "", "Category ID to place channel under")
	channelAddCmd.Flags().BoolVarP(&channelAddFlags.yes, "yes", "y", false, "Skip confirmation prompt")

	channelUpdateCmd.Flags().StringVar(&channelUpdateFlags.channel, "channel", "", "Channel ID to update (required)")
	channelUpdateCmd.Flags().StringVar(&channelUpdateFlags.name, "name", "", "New channel name")
	channelUpdateCmd.Flags().StringVar(&channelUpdateFlags.topic, "topic", "", "Channel topic (text channels only)")
	channelUpdateCmd.Flags().StringVar(&channelUpdateFlags.category, "category", "", "Category ID to move channel to")
	channelUpdateCmd.Flags().BoolVar(&channelUpdateFlags.nsfw, "nsfw", false, "Mark channel as NSFW")
	channelUpdateCmd.Flags().BoolVarP(&channelUpdateFlags.yes, "yes", "y", false, "Skip confirmation prompt")

	channelDeleteCmd.Flags().StringVar(&channelDeleteFlags.channel, "channel", "", "Channel ID to delete (required)")
	channelDeleteCmd.Flags().BoolVarP(&channelDeleteFlags.yes, "yes", "y", false, "Skip confirmation prompt")

	channelMoveCmd.Flags().StringVar(&channelMoveFlags.server, "server", "", "Server ID (defaults to configured server)")
	channelMoveCmd.Flags().StringVar(&channelMoveFlags.channel, "channel", "", "Channel ID to move (required)")
	channelMoveCmd.Flags().IntVar(&channelMoveFlags.position, "position", -1, "New position within the category (0-indexed)")
	channelMoveCmd.Flags().StringVar(&channelMoveFlags.category, "category", "", "Category ID to move channel to (use 'none' to remove from category)")
	channelMoveCmd.Flags().BoolVarP(&channelMoveFlags.yes, "yes", "y", false, "Skip confirmation prompt")

	channelExportCmd.Flags().StringVar(&channelExportFlags.server, "server", "", "Server ID (defaults to configured server)")
	channelExportCmd.Flags().StringVar(&channelExportFlags.file, "file", "", "Output file (defaults to channels.json)")

	channelImportCmd.Flags().StringVar(&channelImportFlags.server, "server", "", "Server ID (defaults to configured server)")
	channelImportCmd.Flags().StringVar(&channelImportFlags.file, "file", "", "JSON file to import (defaults to channels.json)")
	channelImportCmd.Flags().BoolVar(&channelImportFlags.deleteMissing, "delete-missing", false, "Delete channels on the server not present in the file (non-reversible)")
	channelImportCmd.Flags().BoolVarP(&channelImportFlags.yes, "yes", "y", false, "Skip the 'apply' confirmation prompt")
}
