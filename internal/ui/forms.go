package ui

import (
	"errors"

	"github.com/charmbracelet/huh"
)

// RoleFormResult holds the values collected by RoleForm.
type RoleFormResult struct {
	Name         string
	Color        string
	Hoist        bool
	Mentionable  bool
	Permissions  []string
}

// RoleForm renders an interactive form for creating a role, prefilled with the
// given flag values. Returns nil if the user aborts.
func RoleForm(name, color string, hoist, mentionable bool, permissions, permNames []string) (*RoleFormResult, error) {
	r := &RoleFormResult{
		Name:        name,
		Color:       color,
		Hoist:       hoist,
		Mentionable: mentionable,
		Permissions: permissions,
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Role name").Value(&r.Name).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Color (hex)").Description("e.g. FF0000; leave empty for default").Value(&r.Color),
		huh.NewConfirm().Title("Hoist").Description("Display separately in the member list").Affirmative("Yes").Negative("No").Value(&r.Hoist),
		huh.NewConfirm().Title("Mentionable").Description("Allow anyone to mention this role").Affirmative("Yes").Negative("No").Value(&r.Mentionable),
		huh.NewMultiSelect[string]().Title("Permissions").Options(permOptions(permNames)...).Value(&r.Permissions).Height(10),
	))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, nil
		}
		return nil, err
	}
	return r, nil
}

// ChannelFormResult holds the values collected by ChannelForm.
type ChannelFormResult struct {
	Name       string
	Type       string // "text" or "voice"
	CategoryID string
}

// ChannelForm renders an interactive form for creating a channel. categories
// may be nil. Returns nil if the user aborts.
func ChannelForm(name, typ, categoryID string, categories []CategoryOption) (*ChannelFormResult, error) {
	if typ == "" {
		typ = "text"
	}
	r := &ChannelFormResult{Name: name, Type: typ, CategoryID: categoryID}

	typeOpts := []huh.Option[string]{
		huh.NewOption("Text", "text"),
		huh.NewOption("Voice", "voice"),
	}
	catOpts := []huh.Option[string]{huh.NewOption("(No category)", "")}
	for _, c := range categories {
		catOpts = append(catOpts, huh.NewOption(c.Name, c.ID))
	}

	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Channel name").Value(&r.Name).Validate(huh.ValidateNotEmpty()),
		huh.NewSelect[string]().Title("Type").Options(typeOpts...).Value(&r.Type),
		huh.NewSelect[string]().Title("Category").Options(catOpts...).Value(&r.CategoryID),
	))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, nil
		}
		return nil, err
	}
	return r, nil
}

// Overwrite describes one role-based permission overwrite on a channel.
type Overwrite struct {
	RoleID   string
	RoleName string
	Allow    []string
	Deny     []string
}

// OverwritesForm interactively collects role-based permission overwrites. The
// user picks a role, then Allow and Deny permissions, and can add more roles.
// prefill seeds the selections for roles already configured via --allow/--deny.
// Returns nil if the user aborts.
func OverwritesForm(roles []RoleOption, permNames []string, prefill []Overwrite) ([]Overwrite, error) {
	perms := permOptions(permNames)

	// Map prefill by role ID so selections can be seeded.
	allowByRole := map[string][]string{}
	denyByRole := map[string][]string{}
	for _, ow := range prefill {
		allowByRole[ow.RoleID] = ow.Allow
		denyByRole[ow.RoleID] = ow.Deny
	}

	var out []Overwrite
	for {
		roleOpts := []huh.Option[string]{huh.NewOption("Done adding overwrites", "")}
		for _, r := range roles {
			roleOpts = append(roleOpts, huh.NewOption(r.Name, r.ID))
		}

		var roleID string
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Title("Add permission overwrite for role").Options(roleOpts...).Value(&roleID),
		))
		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil, nil
			}
			return nil, err
		}
		if roleID == "" {
			break
		}

		allow := allowByRole[roleID]
		deny := denyByRole[roleID]
		roleName := ""
		for _, r := range roles {
			if r.ID == roleID {
				roleName = r.Name
				break
			}
		}

		owForm := huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().Title("Allow permissions").Description(roleName).Options(perms...).Value(&allow).Height(10),
			huh.NewMultiSelect[string]().Title("Deny permissions").Description(roleName).Options(perms...).Value(&deny).Height(10),
		))
		if err := owForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil, nil
			}
			return nil, err
		}

		out = append(out, Overwrite{RoleID: roleID, RoleName: roleName, Allow: allow, Deny: deny})
		delete(allowByRole, roleID)
		delete(denyByRole, roleID)
	}

	// Preserve any prefilled overwrites whose role was not revisited.
	for _, r := range roles {
		if allows, ok := allowByRole[r.ID]; ok || denyByRole[r.ID] != nil {
			out = append(out, Overwrite{RoleID: r.ID, RoleName: r.Name, Allow: allows, Deny: denyByRole[r.ID]})
		}
	}

	return out, nil
}

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
