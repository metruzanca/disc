# disc command reference

Full documentation for every `disc` command. See the [README](../README.md) for installation and configuration.

All server commands accept `--server <server-id>` to override the default server. Mutating commands ask for confirmation unless `--yes` is given.

## Status

Show connection status and the servers the bot belongs to.

```bash
disc status
```

## Invite

Generate an OAuth2 invite link to add the bot to a server.

```bash
disc invite
```

## Channels

```bash
# List channels, grouped by category
disc channel list

# Create a channel (text by default)
disc channel add --name general
disc channel add --name voice-chat --type voice
disc channel add --name help --category 987654321

# Update a channel
disc channel update --channel 123456789 --name new-name
disc channel update --channel 123456789 --topic "Welcome to the channel"
disc channel update --channel 123456789 --category 987654321

# Delete a channel
disc channel delete --channel 123456789

# Move a channel to a position or category
disc channel move --channel 111111111 --position 0
disc channel move --channel 111111111 --category 222222222
disc channel move --channel 111111111 --category 222222222 --position 3

# Export the channel structure to JSON (channels.json by default)
disc channel export
disc channel export --file channels.json

# Import a channel definition (dry run first; type "apply" to proceed)
disc channel import
disc channel import --file channels.json
# Also delete channels not present in the file (non-reversible)
disc channel import --file channels.json --delete-missing
```

- `disc channel add` — `--name` (required), `--type` (`text` or `voice`, default `text`), `--category`
- `disc channel update` — `--channel` (required), `--name`, `--topic`, `--category`, `--nsfw`
- `disc channel delete` — `--channel` (required)
- `disc channel move` — `--channel` (required), `--position` (0-indexed), `--category` (use `none` to remove from a category)
- `disc channel export` — `--file` (defaults to `channels.json`)
- `disc channel import` — `--file` (defaults to `channels.json`), `--delete-missing`, `--yes`

Channels are matched by name on import. Channels already present with the
correct settings are left unchanged. System channels (rules, updates, welcome)
are never deleted. With `--delete-missing`, categories are only removed when
none of their child channels are being kept.

## Roles

```bash
# List roles, sorted by hierarchy (highest first)
disc role list

# Show role details
disc role show --role 987654321

# Create a role
disc role add --name "Moderator"
disc role add --name "VIP" --color FF0000 --hoist
disc role add --name "Helper" --mentionable

# Update a role
disc role update --role 987654321 --name "New Name"
disc role update --role 987654321 --color 00FF00
disc role update --role 987654321 --hoist=false

# Delete a role
disc role delete --role 987654321

# Export roles to JSON (roles.json by default)
disc role export
disc role export --file roles.json

# Import a role definition (dry run first; type "apply" to proceed)
disc role import
disc role import --file roles.json
# Also delete roles not present in the file (non-reversible)
disc role import --file roles.json --delete-missing
```

- `disc role add` — `--name` (required), `--color` (hex, e.g. `FF0000`), `--hoist`, `--mentionable`
- `disc role update` — `--role` (required), `--name`, `--color`, `--hoist`, `--mentionable`
- `disc role delete` — `--role` (required)
- `disc role export` — `--file` (defaults to `roles.json`)
- `disc role import` — `--file` (defaults to `roles.json`), `--delete-missing`, `--yes`

Roles are matched by name on import. Roles already present with the correct
settings (color, hoist, mentionable, permissions) are left unchanged. Managed
(integration/bot) roles and `@everyone` are never deleted.

## Scheduled events

```bash
# List scheduled events (add --active to only show scheduled/active)
disc event list
disc event list --active

# Show event details
disc event show --event 987654321

# Create a voice event in a channel
disc event add --name "Coffee Break" --start "2026-01-15 19:00" --channel 123456789

# Create an external event (requires --location and --end)
disc event add --name "Meetup" --type external --location "Local Cafe" \
  --start "2026-01-15 19:00" --end "2026-01-15 21:00"

# Update an event (name, description, time, type, or status)
disc event update --event 987654321 --name "New Name"
disc event update --event 987654321 --status active

# Delete an event
disc event delete --event 987654321
```

- `disc event add` — `--name` (required), `--start` (required), `--channel`, `--description`, `--type`, `--location`, `--end`
- `disc event update` — `--event` (required), `--name`, `--description`, `--channel`, `--start`, `--end`, `--location`, `--type`, `--status`
- `disc event delete` — `--event` (required)

### Event types

- `voice` (default) — hosted in a voice channel
- `stage` — hosted in a stage channel
- `external` — hosted outside Discord; requires a `--location` and an `--end` time

### Event status

Use `--status` on update to move an event through its lifecycle: `scheduled`, `active`, `completed`, `canceled`.
