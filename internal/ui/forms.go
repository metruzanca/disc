package ui

import (
	"errors"

	"github.com/charmbracelet/huh"
)

// EventFormResult holds the values collected by EventForm.
type EventFormResult struct {
	Name        string
	Description string
	Type        string // voice, stage, or external
	Start       string
	End         string
	Location    string
	ChannelID   string
}

// EventForm renders an interactive form for creating a scheduled event. Fields
// are shown conditionally based on the chosen type. Returns nil if the user
// aborts.
func EventForm(name, description, etype, start, end, location, channelID string, channels []ChannelOption) (*EventFormResult, error) {
	if etype == "" {
		etype = "voice"
	}
	r := &EventFormResult{
		Name:        name,
		Description: description,
		Type:        etype,
		Start:       start,
		End:         end,
		Location:    location,
		ChannelID:   channelID,
	}

	typeOpts := []huh.Option[string]{
		huh.NewOption("Voice", "voice"),
		huh.NewOption("Stage", "stage"),
		huh.NewOption("External", "external"),
	}

	chanOpts := []huh.Option[string]{huh.NewOption("(Default)", "")}
	for _, c := range channels {
		chanOpts = append(chanOpts, huh.NewOption(c.Name, c.ID))
	}

	inGuild := func() bool { return r.Type != "external" }

	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Event name").Value(&r.Name).Validate(huh.ValidateNotEmpty()),
		huh.NewSelect[string]().Title("Type").Options(typeOpts...).Value(&r.Type),
		huh.NewText().Title("Description").Description("Optional").Value(&r.Description).CharLimit(1000).Lines(3),
		huh.NewInput().Title("Start time").Description("e.g. 2006-01-02 15:04 or RFC3339").Value(&r.Start).Validate(huh.ValidateNotEmpty()),
		onlyWhen(huh.NewInput().Title("End time").Description("Required for external events").Value(&r.End), func() bool { return r.Type == "external" }),
		onlyWhen(huh.NewInput().Title("Location").Description("Required for external events").Value(&r.Location), func() bool { return r.Type == "external" }),
		onlyWhen(huh.NewSelect[string]().Title("Channel").Options(chanOpts...).Value(&r.ChannelID), inGuild),
	))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, nil
		}
		return nil, err
	}
	return r, nil
}
