# disc command reference

Full documentation for every `disc` command. See the [README](../README.md) for installation and configuration.

All server commands accept `--server <server-id>` to override the default server. Mutating commands ask for confirmation unless `--yes` is given; `--dry` prints what would happen and exits without changing anything; `--agent` forces non-interactive behavior (erroring on missing required params) for scripts and agents.

In an interactive terminal, the `add` commands (`channel add`, `role add`, `event add`) show a form prefilled from any flags passed, with multi-select pickers for permissions.

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

## Channels (read-only)

```bash
# List channels, grouped by category
disc channel list

# Show channel details, including permission overwrites
disc channel show --channel 123456789
```

## Roles (read-only)

```bash
# List roles, sorted by hierarchy (highest first)
disc role list

# Show role details
disc role show --role 987654321

# List the exact names to use with --permissions
disc role perm list
```

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
- `disc config push` — `--server`, `--file`, `--delete-missing`, `--yes`, `--dry`, `--agent`
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
- All mutating event commands accept `--yes`, `--dry`, and `--agent`.

### Event types

- `voice` (default) — hosted in a voice channel
- `stage` — hosted in a stage channel
- `external` — hosted outside Discord; requires a `--location` and an `--end` time

### Event status

Use `--status` on update to move an event through its lifecycle: `scheduled`, `active`, `completed`, `canceled`.
