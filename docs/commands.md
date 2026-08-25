# disc command reference

Full documentation for every `disc` command. See the [README](../README.md) for installation and configuration.

All server commands accept `--server <server-id>` to override the active server.
Mutating commands ask for confirmation unless `--yes` is given; `--dry` prints
what would happen and exits without changing anything; `--agent` forces
non-interactive behavior (erroring on missing required params) for scripts and agents.

In an interactive terminal, `disc event add` shows a form prefilled from any flags passed.

## Server management

```bash
# Add a new bot and server to disc
disc server new

# Switch the active server
disc server switch

# List all managed bots and servers
disc server list

# Snapshot the live server into the config file
disc server pull
disc server pull --file server.json

# Apply the config to the live server (shows a dry run first, then confirms)
disc server push
disc server push --delete-missing
# Non-interactive: preview with --dry, apply with --yes
disc server push --dry
disc server push --yes
disc server push --delete-missing --yes
```

- `disc server new` — `--token`, `--config-location` (`local`|`global`), `--server-id` (required when bot is in multiple servers)
- `disc server switch` — `--server` (non-interactive); interactive with no flags
- `disc server list` — no flags
- `disc server pull` — `--file`
- `disc server push` — `--file`, `--delete-missing`, `--yes`, `--dry`, `--agent`

## Status

Show all bots and servers under disc management. Connects to each bot to list
its guilds, filtering to managed servers and marking active ones.

```bash
disc status
```

## Invite

Generate an OAuth2 invite link to add the active bot to a server.

```bash
disc invite
```

## Channels

```bash
# List channels, grouped by category
disc channel list

# Show channel details, including permission overwrites
disc channel show --channel 123456789

# Add a channel to the config (not created on server until push)
disc channel add --name general
disc channel add --name help --category Staff
disc channel add --name voice-chat --type voice

# Update a channel in the config
disc channel update --channel 123456789 --name new-name
disc channel update --channel 123456789 --topic "Welcome"
disc channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"

# Delete a channel from the config
disc channel delete --channel 123456789

# Move a channel within the config
disc channel move --channel 111111111 --position 0
disc channel move --channel 111111111 --category Staff --position 3
```

- `disc channel list` — `--server`
- `disc channel show` — `--server`, `--channel` (required)
- `disc channel add` — `--file`, `--name` (required), `--type` (`text`|`voice`|`category`|`news`|`stage`|`forum`), `--topic`, `--nsfw`, `--category`, `--position`, `--allow`, `--deny`
- `disc channel update` — `--file`, `--channel` (required), `--name`, `--topic`, `--nsfw`, `--category`, `--allow`, `--deny`, `--no-overwrite`
- `disc channel delete` — `--file`, `--channel` (required)
- `disc channel move` — `--file`, `--channel` (required), `--position`, `--category`

`--allow` and `--deny` take `Role:Perm,Perm` values and are repeatable. System channels (rules, updates, welcome) are never deleted by `disc server push`.

## Roles (read-only)

```bash
# List roles, sorted by hierarchy (highest first)
disc role list

# Show role details
disc role show --role 987654321

# List the exact names to use with --permissions
disc role perm list
```

## Role config editing

Roles are edited in the config file (not created on server until push).

```bash
disc role add --name Moderator --permissions "Kick Members,Ban Members"
disc role add --name VIP --color FF0000 --hoist
disc role update --role 987654321 --name "New Name"
disc role update --role 987654321 --color 00FF00
disc role update --role 987654321 --position 3
disc role delete --role 987654321
```

- `disc role add` — `--file`, `--name` (required), `--color`, `--hoist`, `--mentionable`, `--permissions`
- `disc role update` — `--file`, `--role` (required), `--name`, `--color`, `--hoist`, `--mentionable`, `--permissions`, `--position`
- `disc role delete` — `--file`, `--role` (required)

Managed (integration/bot) roles and `@everyone` are never deleted by `disc server push`.

## Channel config editing

Channels are edited in the config file (not created on server until push). Permissions use `Role:Perm,Perm` syntax.

See channel subcommands above.

## Scheduled events

```bash
# List scheduled events (add --active to only show scheduled/active)
disc event list
disc event list --active

# Show event details
disc event show --event 987654321

# Create a voice event in a channel (default type)
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
