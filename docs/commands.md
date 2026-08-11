# disc command reference

Full documentation for every `disc` command. See the [README](../README.md) for installation and configuration.

All server commands accept `--server <server-id>` to override the default server. Mutating commands ask for confirmation unless `--yes` is given; `--dry` prints what would happen and exits without changing anything.

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

# Show channel details, including permission overwrites
disc channel show --channel 123456789

# Set permission overwrites (role-based) on create or update
disc channel add --name general --allow "Moderator:Send Messages"
disc channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
```

- `disc channel add` — `--name` (required), `--type` (`text` or `voice`, default `text`), `--category`, `--allow`, `--deny`
- `disc channel update` — `--channel` (required), `--name`, `--topic`, `--category`, `--nsfw`, `--allow`, `--deny`
- `disc channel delete` — `--channel` (required)
- `disc channel move` — `--channel` (required), `--position` (0-indexed), `--category` (use `none` to remove from a category)
- `disc channel show` — `--channel` (required)
- All mutating channel commands accept `--yes` (apply without prompting) and `--dry` (show what would happen only).

`--allow` and `--deny` take `Role:Perm,Perm` values and are repeatable; the
un-provided side of an overwrite is preserved. Member-level overwrites are not
managed by disc. System channels (rules, updates, welcome) are never deleted by
`disc config push`.

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

# List the exact names to use with --permissions
disc role perm list
```

- `disc role add` — `--name` (required), `--color` (hex, e.g. `FF0000`), `--hoist`, `--mentionable`, `--permissions`
- `disc role update` — `--role` (required), `--name`, `--color`, `--hoist`, `--mentionable`, `--permissions`
- `disc role delete` — `--role` (required)
- `disc role perm list` — lists every recognized permission name
- All mutating role commands accept `--yes` and `--dry`.

Managed (integration/bot) roles and `@everyone` are never deleted by
`disc config push`.

## Declarative config

Manage roles and channels as declarative JSON in a single `server.json`
file instead of applying changes to the server directly.

```bash
# Snapshot the live server into server.json (roles + channels together)
disc config pull
disc config pull --server 123456789
disc config pull --file server.json

# Edit the config file declaratively (applies on the next push)
disc config role add --name Moderator --permissions "Kick Members,Ban Members"
disc config role update --role 987654321 --name "New Name"
disc config role update --role 987654321 --color 00FF00
disc config role update --role 987654321 --position 3
disc config role delete --role 987654321
disc config channel add --name help --category Staff
disc config channel update --channel 123456789 --topic "Welcome to the channel"
disc config channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
disc config channel update --channel 123456789 --no-overwrite Moderator
disc config channel move --channel 111111111 --position 0
disc config channel delete --channel 123456789

# Apply the config to the live server (shows a dry run first, then confirms)
disc config push
disc config push --file server.json
disc config push --delete-missing
# Non-interactive: preview with --dry, apply with --yes
disc config push --dry
disc config push --yes
disc config push --delete-missing --yes
```

- `disc config pull` — `--server`, `--file` (defaults to `server.json`)
- `disc config push` — `--server`, `--file`, `--delete-missing`, `--yes`, `--dry`
- `disc config role add` — `--name` (required), `--color`, `--hoist`, `--mentionable`, `--permissions`
- `disc config role update` — `--role` (required), `--name`, `--color`, `--hoist`, `--mentionable`, `--permissions`, `--position`
- `disc config role delete` — `--role` (required)
- `disc config channel add` — `--name` (required), `--type`, `--topic`, `--nsfw`, `--category`, `--position`, `--allow`, `--deny`
- `disc config channel update` — `--channel` (required), `--name`, `--topic`, `--nsfw`, `--category`, `--allow`, `--deny`, `--no-overwrite`
- `disc config channel delete` — `--channel` (required)
- `disc config channel move` — `--channel` (required), `--position`, `--category`

`disc config pull` stores each role and channel's server ID, their positions,
and role-based channel permission overwrites. Edit commands reference resources
by ID (`--role` / `--channel`), matching the server CLI. Entries added via
`add` have no ID until they are pushed and pulled. On push, each entry is
matched to the server by ID when present, otherwise by name, so renames update
in place and push is idempotent. Roles are reordered to match their configured
positions. With `--delete-missing`, categories are only removed when none of
their child channels are being kept, and live role overwrites absent from the
config are removed (member overwrites are always preserved). Managed roles,
`@everyone`, and system channels are never deleted.

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

# Copy an event, inheriting all properties (override any with the add flags)
disc event copy --event 987654321
disc event copy --event 987654321 --name "Coffee Break (Week 2)" --start "2026-01-22 19:00"
```

- `disc event add` — `--name` (required), `--start` (required), `--channel`, `--description`, `--type`, `--location`, `--end`
- `disc event update` — `--event` (required), `--name`, `--description`, `--channel`, `--start`, `--end`, `--location`, `--type`, `--status`
- `disc event delete` — `--event` (required)
- `disc event copy` — `--event` (required); any `add` flag can be used to override that property
- All mutating event commands accept `--yes` and `--dry`.

### Event types

- `voice` (default) — hosted in a voice channel
- `stage` — hosted in a stage channel
- `external` — hosted outside Discord; requires a `--location` and an `--end` time

### Event status

Use `--status` on update to move an event through its lifecycle: `scheduled`, `active`, `completed`, `canceled`.
