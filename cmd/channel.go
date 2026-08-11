package cmd

import (
	"fmt"
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

var serverFlag string

func init() {
	channelCmd.AddCommand(channelListCmd)
	channelCmd.AddCommand(channelAddCmd)
	channelCmd.AddCommand(channelUpdateCmd)
	channelCmd.AddCommand(channelDeleteCmd)
	channelCmd.AddCommand(channelMoveCmd)

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
}
