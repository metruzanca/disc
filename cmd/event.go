package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/metruzanca/disc/internal/ui"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Manage scheduled events",
	Long:  "Commands for listing and managing scheduled events in a server.",
}

var (
	eventListFlags = struct {
		server string
		active bool
	}{}
)

var eventListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scheduled events in a server",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(eventListFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		events, err := client.Session().GuildScheduledEvents(serverID, true)
		if err != nil {
			return fmt.Errorf("failed to list events: %w", err)
		}

		// Sort by start time, soonest first.
		sort.Slice(events, func(i, j int) bool {
			return events[i].ScheduledStartTime.Before(events[j].ScheduledStartTime)
		})

		if eventListFlags.active {
			filtered := events[:0]
			for _, e := range events {
				if e.Status == discordgo.GuildScheduledEventStatusScheduled ||
					e.Status == discordgo.GuildScheduledEventStatusActive {
					filtered = append(filtered, e)
				}
			}
			events = filtered
		}

		if len(events) == 0 {
			util.Yellow.Println("No scheduled events.")
			return nil
		}

		for _, e := range events {
			fmt.Printf("  %s ", e.Name)
			util.Dim.Printf("(ID: %s)\n", e.ID)
			util.Cyan.Printf("    Start: %s\n", e.ScheduledStartTime.Format(time.RFC3339))
			if e.ScheduledEndTime != nil {
				util.Cyan.Printf("    End:   %s\n", e.ScheduledEndTime.Format(time.RFC3339))
			}
			util.Cyan.Printf("    Status: %s\n", eventStatusName(e.Status))
			if e.UserCount > 0 {
				util.Cyan.Printf("    Subscribers: %d\n", e.UserCount)
			}
			fmt.Println()
		}
		return nil
	},
}

var (
	eventShowFlags = struct {
		server string
		event  string
	}{}
)

var eventShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show scheduled event details",
	Example: `  disc event show --event 987654321`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if eventShowFlags.event == "" {
			return fmt.Errorf("--event is required")
		}
		client, serverID, err := newClientAndServer(eventShowFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		e, err := client.Session().GuildScheduledEvent(serverID, eventShowFlags.event, true)
		if err != nil {
			return fmt.Errorf("failed to load event: %w", err)
		}

		util.Bold.Printf("Event: %s\n", e.Name)
		util.Cyan.Printf("ID: %s\n", e.ID)
		if e.Description != "" {
			util.Cyan.Printf("Description: %s\n", e.Description)
		}
		util.Cyan.Printf("Start: %s\n", e.ScheduledStartTime.Format(time.RFC3339))
		if e.ScheduledEndTime != nil {
			util.Cyan.Printf("End: %s\n", e.ScheduledEndTime.Format(time.RFC3339))
		}
		util.Cyan.Printf("Status: %s\n", eventStatusName(e.Status))
		util.Cyan.Printf("Type: %s\n", eventEntityTypeName(e.EntityType))
		if e.ChannelID != "" {
			util.Cyan.Printf("Channel: %s\n", e.ChannelID)
		}
		if e.EntityMetadata.Location != "" {
			util.Cyan.Printf("Location: %s\n", e.EntityMetadata.Location)
		}
		if e.UserCount > 0 {
			util.Cyan.Printf("Subscribers: %d\n", e.UserCount)
		}
		return nil
	},
}

var (
	eventAddFlags = struct {
		server      string
		name        string
		description string
		channel     string
		start       string
		end         string
		location    string
		entityType  string
		yes         bool
		dry         bool
	}{}
)

var eventAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a scheduled event",
	Long: `Create a scheduled event in a Discord server.

Entity types:
  voice       hosted in a voice channel (default)
  stage       hosted in a stage channel
  external    hosted externally (requires --location and --end)

Examples:
  disc event add --name "Coffee Break" --start "2026-01-15 19:00"
  disc event add --name "Meetup" --type external --location "Local Cafe" --start "2026-01-15 19:00" --end "2026-01-15 21:00"
  disc event add --name "Talk" --type stage --channel 123456789 --start "2026-01-15 19:00"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, serverID, err := newClientAndServer(eventAddFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		if interactive(eventAddFlags.yes, eventAddFlags.dry) {
			channels, err := uiChannels(client.Session(), serverID)
			if err != nil {
				return err
			}
			res, err := ui.EventForm(eventAddFlags.name, eventAddFlags.description, eventAddFlags.entityType, eventAddFlags.start, eventAddFlags.end, eventAddFlags.location, eventAddFlags.channel, channels)
			if err != nil {
				return err
			}
			if res == nil {
				util.Yellow.Println("Aborted.")
				return nil
			}
			eventAddFlags.name = res.Name
			eventAddFlags.description = res.Description
			eventAddFlags.entityType = res.Type
			eventAddFlags.start = res.Start
			eventAddFlags.end = res.End
			eventAddFlags.location = res.Location
			eventAddFlags.channel = res.ChannelID
		}
		if eventAddFlags.name == "" {
			return fmt.Errorf("--name is required")
		}
		if eventAddFlags.start == "" {
			return fmt.Errorf("--start is required")
		}

		start, err := util.ParseTime(eventAddFlags.start)
		if err != nil {
			return fmt.Errorf("invalid --start: %w", err)
		}

		entityType, err := parseEntityType(eventAddFlags.entityType)
		if err != nil {
			return err
		}

		var end *time.Time
		if eventAddFlags.end != "" {
			t, err := util.ParseTime(eventAddFlags.end)
			if err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}
			end = &t
		}

		if entityType == discordgo.GuildScheduledEventEntityTypeExternal {
			if eventAddFlags.location == "" {
				return fmt.Errorf("--location is required for external events")
			}
			if end == nil {
				return fmt.Errorf("--end is required for external events")
			}
		}

		summary := fmt.Sprintf("Create event '%s' in server %s?", eventAddFlags.name, serverID)
		proceed, err := confirmRun(summary, eventAddFlags.yes, eventAddFlags.dry)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		params := &discordgo.GuildScheduledEventParams{
			Name:               eventAddFlags.name,
			Description:        eventAddFlags.description,
			ChannelID:          eventAddFlags.channel,
			ScheduledStartTime: &start,
			ScheduledEndTime:   end,
			PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
			EntityType:         entityType,
		}
		if eventAddFlags.location != "" {
			params.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{
				Location: eventAddFlags.location,
			}
		}

		e, err := client.Session().GuildScheduledEventCreate(serverID, params)
		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}
		util.Green.Printf("Created event %s (%s)\n", e.Name, e.ID)
		return nil
	},
}

var (
	eventUpdateFlags = struct {
		server      string
		event       string
		name        string
		description string
		channel     string
		start       string
		end         string
		location    string
		entityType  string
		status      string
		yes         bool
		dry         bool
	}{}
)

var eventUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a scheduled event",
	Long: `Update an existing scheduled event's properties.

Use --status to start, end, or cancel an event (scheduled, active, completed, canceled).

Examples:
  disc event update --event 987654321 --name "New Name"
  disc event update --event 987654321 --status active`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if eventUpdateFlags.event == "" {
			return fmt.Errorf("--event is required")
		}
		client, serverID, err := newClientAndServer(eventUpdateFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		proceed, err := confirmRun(fmt.Sprintf("Update event %s?", eventUpdateFlags.event), eventUpdateFlags.yes, eventUpdateFlags.dry)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		params := &discordgo.GuildScheduledEventParams{}

		if eventUpdateFlags.name != "" {
			params.Name = eventUpdateFlags.name
		}
		if cmd.Flags().Changed("description") {
			params.Description = eventUpdateFlags.description
		}
		if cmd.Flags().Changed("channel") {
			params.ChannelID = eventUpdateFlags.channel
		}
		if eventUpdateFlags.start != "" {
			t, err := util.ParseTime(eventUpdateFlags.start)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			params.ScheduledStartTime = &t
		}
		if eventUpdateFlags.end != "" {
			t, err := util.ParseTime(eventUpdateFlags.end)
			if err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}
			params.ScheduledEndTime = &t
		}
		if eventUpdateFlags.location != "" {
			params.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{
				Location: eventUpdateFlags.location,
			}
		}
		if eventUpdateFlags.entityType != "" {
			et, err := parseEntityType(eventUpdateFlags.entityType)
			if err != nil {
				return err
			}
			params.EntityType = et
		}
		if eventUpdateFlags.status != "" {
			st, err := parseEventStatus(eventUpdateFlags.status)
			if err != nil {
				return err
			}
			params.Status = st
		}

		e, err := client.Session().GuildScheduledEventEdit(serverID, eventUpdateFlags.event, params)
		if err != nil {
			return fmt.Errorf("failed to update event: %w", err)
		}
		util.Green.Printf("Updated event %s (%s)\n", e.Name, e.ID)
		return nil
	},
}

var (
	eventDeleteFlags = struct {
		server string
		event  string
		yes    bool
		dry    bool
	}{}
)

var eventDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a scheduled event",
	RunE: func(cmd *cobra.Command, args []string) error {
		if eventDeleteFlags.event == "" {
			return fmt.Errorf("--event is required")
		}
		client, serverID, err := newClientAndServer(eventDeleteFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		proceed, err := confirmRun(fmt.Sprintf("Delete event %s?", eventDeleteFlags.event), eventDeleteFlags.yes, eventDeleteFlags.dry)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		if err := client.Session().GuildScheduledEventDelete(serverID, eventDeleteFlags.event); err != nil {
			return fmt.Errorf("failed to delete event: %w", err)
		}
		util.Green.Printf("Deleted event %s\n", eventDeleteFlags.event)
		return nil
	},
}

var (
	eventCopyFlags = struct {
		server      string
		event       string
		name        string
		description string
		channel     string
		start       string
		end         string
		location    string
		entityType  string
		yes         bool
		dry         bool
	}{}
)

var eventCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy a scheduled event",
	Long: `Copy an existing scheduled event to create a new one, inheriting all of its
properties. Use the same flags as "add" to override any property on the new
event. The new event is always created as scheduled.

Examples:
  disc event copy --event 987654321
  disc event copy --event 987654321 --name "Coffee Break (Week 2)" --start "2026-01-22 19:00"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if eventCopyFlags.event == "" {
			return fmt.Errorf("--event is required")
		}
		client, serverID, err := newClientAndServer(eventCopyFlags.server)
		if err != nil {
			return err
		}
		defer client.Close()

		src, err := client.Session().GuildScheduledEvent(serverID, eventCopyFlags.event, true)
		if err != nil {
			return fmt.Errorf("failed to load source event: %w", err)
		}

		params := &discordgo.GuildScheduledEventParams{
			Name:         src.Name,
			Description:  src.Description,
			PrivacyLevel: discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
			EntityType:   src.EntityType,
			ChannelID:    src.ChannelID,
		}

		if eventCopyFlags.name != "" {
			params.Name = eventCopyFlags.name
		}
		if cmd.Flags().Changed("description") {
			params.Description = eventCopyFlags.description
		}
		if cmd.Flags().Changed("channel") {
			params.ChannelID = eventCopyFlags.channel
		}

		start := src.ScheduledStartTime
		if eventCopyFlags.start != "" {
			t, err := util.ParseTime(eventCopyFlags.start)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			start = t
		}
		params.ScheduledStartTime = &start

		if cmd.Flags().Changed("end") {
			if eventCopyFlags.end == "" {
				params.ScheduledEndTime = nil
			} else {
				t, err := util.ParseTime(eventCopyFlags.end)
				if err != nil {
					return fmt.Errorf("invalid --end: %w", err)
				}
				params.ScheduledEndTime = &t
			}
		} else {
			params.ScheduledEndTime = src.ScheduledEndTime
		}

		location := src.EntityMetadata.Location
		if cmd.Flags().Changed("location") {
			location = eventCopyFlags.location
		}
		if cmd.Flags().Changed("type") {
			et, err := parseEntityType(eventCopyFlags.entityType)
			if err != nil {
				return err
			}
			params.EntityType = et
		}

		if params.EntityType == discordgo.GuildScheduledEventEntityTypeExternal {
			if location == "" {
				return fmt.Errorf("--location is required for external events")
			}
			if params.ScheduledEndTime == nil {
				return fmt.Errorf("--end is required for external events")
			}
		}
		if location != "" {
			params.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{Location: location}
		}

		// A copy is always created as a fresh scheduled event.
		params.Status = discordgo.GuildScheduledEventStatusScheduled

		summary := fmt.Sprintf("Create copy of event '%s' as '%s'?", src.Name, params.Name)
		proceed, err := confirmRun(summary, eventCopyFlags.yes, eventCopyFlags.dry)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		e, err := client.Session().GuildScheduledEventCreate(serverID, params)
		if err != nil {
			return fmt.Errorf("failed to copy event: %w", err)
		}
		util.Green.Printf("Created event %s (%s)\n", e.Name, e.ID)
		return nil
	},
}

func parseEntityType(s string) (discordgo.GuildScheduledEventEntityType, error) {
	switch strings.ToLower(s) {
	case "", "voice":
		return discordgo.GuildScheduledEventEntityTypeVoice, nil
	case "stage":
		return discordgo.GuildScheduledEventEntityTypeStageInstance, nil
	case "external":
		return discordgo.GuildScheduledEventEntityTypeExternal, nil
	default:
		return 0, fmt.Errorf("invalid entity type '%s'; must be voice, stage, or external", s)
	}
}

func parseEventStatus(s string) (discordgo.GuildScheduledEventStatus, error) {
	switch strings.ToLower(s) {
	case "scheduled":
		return discordgo.GuildScheduledEventStatusScheduled, nil
	case "active":
		return discordgo.GuildScheduledEventStatusActive, nil
	case "completed":
		return discordgo.GuildScheduledEventStatusCompleted, nil
	case "canceled", "cancelled":
		return discordgo.GuildScheduledEventStatusCanceled, nil
	default:
		return 0, fmt.Errorf("invalid status '%s'; must be scheduled, active, completed, or canceled", s)
	}
}

func eventStatusName(s discordgo.GuildScheduledEventStatus) string {
	switch s {
	case discordgo.GuildScheduledEventStatusScheduled:
		return "scheduled"
	case discordgo.GuildScheduledEventStatusActive:
		return "active"
	case discordgo.GuildScheduledEventStatusCompleted:
		return "completed"
	case discordgo.GuildScheduledEventStatusCanceled:
		return "canceled"
	default:
		return fmt.Sprintf("unknown (%d)", int(s))
	}
}

func eventEntityTypeName(t discordgo.GuildScheduledEventEntityType) string {
	switch t {
	case discordgo.GuildScheduledEventEntityTypeStageInstance:
		return "stage"
	case discordgo.GuildScheduledEventEntityTypeVoice:
		return "voice"
	case discordgo.GuildScheduledEventEntityTypeExternal:
		return "external"
	default:
		return fmt.Sprintf("unknown (%d)", int(t))
	}
}

func init() {
	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventShowCmd)
	eventCmd.AddCommand(eventAddCmd)
	eventCmd.AddCommand(eventUpdateCmd)
	eventCmd.AddCommand(eventDeleteCmd)
	eventCmd.AddCommand(eventCopyCmd)

	eventListCmd.Flags().StringVar(&eventListFlags.server, "server", "", "Server ID (defaults to configured server)")
	eventListCmd.Flags().BoolVar(&eventListFlags.active, "active", false, "Only show scheduled and active events")

	eventShowCmd.Flags().StringVar(&eventShowFlags.server, "server", "", "Server ID (defaults to configured server)")
	eventShowCmd.Flags().StringVar(&eventShowFlags.event, "event", "", "Event ID to show details for (required)")

	eventAddCmd.Flags().StringVar(&eventAddFlags.server, "server", "", "Server ID (defaults to configured server)")
	eventAddCmd.Flags().StringVar(&eventAddFlags.name, "name", "", "Event name (required)")
	eventAddCmd.Flags().StringVar(&eventAddFlags.description, "description", "", "Event description")
	eventAddCmd.Flags().StringVar(&eventAddFlags.channel, "channel", "", "Channel ID to host the event in")
	eventAddCmd.Flags().StringVar(&eventAddFlags.start, "start", "", "Start time (e.g. 2006-01-02 15:04 or RFC3339)")
	eventAddCmd.Flags().StringVar(&eventAddFlags.end, "end", "", "End time (required for external events)")
	eventAddCmd.Flags().StringVar(&eventAddFlags.location, "location", "", "Location (required for external events)")
	eventAddCmd.Flags().StringVar(&eventAddFlags.entityType, "type", "voice", "Entity type: voice, stage, or external")
	eventAddCmd.Flags().BoolVarP(&eventAddFlags.yes, "yes", "y", false, "Skip confirmation prompt")
	eventAddCmd.Flags().BoolVar(&eventAddFlags.dry, "dry", false, "Show what would happen without making changes")

	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.server, "server", "", "Server ID (defaults to configured server)")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.event, "event", "", "Event ID to update (required)")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.name, "name", "", "New event name")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.description, "description", "", "New event description")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.channel, "channel", "", "New channel ID")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.start, "start", "", "New start time")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.end, "end", "", "New end time")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.location, "location", "", "New location (external events)")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.entityType, "type", "", "New entity type: voice, stage, or external")
	eventUpdateCmd.Flags().StringVar(&eventUpdateFlags.status, "status", "", "New status: scheduled, active, completed, or canceled")
	eventUpdateCmd.Flags().BoolVarP(&eventUpdateFlags.yes, "yes", "y", false, "Skip confirmation prompt")
	eventUpdateCmd.Flags().BoolVar(&eventUpdateFlags.dry, "dry", false, "Show what would happen without making changes")

	eventDeleteCmd.Flags().StringVar(&eventDeleteFlags.server, "server", "", "Server ID (defaults to configured server)")
	eventDeleteCmd.Flags().StringVar(&eventDeleteFlags.event, "event", "", "Event ID to delete (required)")
	eventDeleteCmd.Flags().BoolVarP(&eventDeleteFlags.yes, "yes", "y", false, "Skip confirmation prompt")
	eventDeleteCmd.Flags().BoolVar(&eventDeleteFlags.dry, "dry", false, "Show what would happen without making changes")

	eventCopyCmd.Flags().StringVar(&eventCopyFlags.server, "server", "", "Server ID (defaults to configured server)")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.event, "event", "", "Event ID to copy (required)")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.name, "name", "", "Event name (defaults to the source's name)")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.description, "description", "", "Event description")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.channel, "channel", "", "Channel ID to host the event in")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.start, "start", "", "Start time (e.g. 2006-01-02 15:04 or RFC3339)")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.end, "end", "", "End time (required for external events)")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.location, "location", "", "Location (required for external events)")
	eventCopyCmd.Flags().StringVar(&eventCopyFlags.entityType, "type", "", "Entity type: voice, stage, or external")
	eventCopyCmd.Flags().BoolVarP(&eventCopyFlags.yes, "yes", "y", false, "Skip confirmation prompt")
	eventCopyCmd.Flags().BoolVar(&eventCopyFlags.dry, "dry", false, "Show what would happen without making changes")
}

// uiChannels returns voice/stage channel options for the event form.
func uiChannels(s *discordgo.Session, serverID string) ([]ui.ChannelOption, error) {
	channels, err := s.GuildChannels(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	var out []ui.ChannelOption
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildVoice || ch.Type == discordgo.ChannelTypeGuildStageVoice {
			out = append(out, ui.ChannelOption{Name: ch.Name, ID: ch.ID, Voice: ch.Type == discordgo.ChannelTypeGuildVoice})
		}
	}
	return out, nil
}
