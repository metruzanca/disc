# disc

A CLI for managing Discord servers via a bot.

Manage Discord servers, channels and roles. Designed for automation and agent-based workflows.

## Features

- **Server discovery** — list channels and roles to see the current structure of a server
- **Channel management** — create, update, delete, and move channels
- **Role management** — create, update, delete, and inspect roles
- **Scheduled event management** — create, update, delete, and inspect scheduled events
- **Automation friendly** — non-interactive via `--yes` for scripted use
- **Flexible config** — token and server ID via environment variables or a local `.env` file

## Requirements

- [Go](https://go.dev/dl/) 1.25+
- A Discord bot token from the [Discord Developer Portal](https://discord.com/developers/applications)
  - Create an application → open the **Bot** tab → **Reset Token** / **Copy** to get your token
  - Grant the bot the **Manage Channels**, **Manage Roles**, **View Channels**, **Send Messages**, **Read Message History**, and **Create Events** permissions when inviting it

## Installation

```bash
go install github.com/metruzanca/disc@latest
```

Or build from source:

```bash
git clone <your-repo-url> disc
cd disc
go build -o disc .
```

The binary is placed in the current directory. To install it into your `GOBIN`:

```bash
go install .
```

## Configuration

`disc` is configured entirely through environment variables, optionally
loaded from a local `.env` file in the directory you run it from. There is
no global config file, so you can run multiple bots with the same `disc`
binary without them stepping on each other — each instance has its own
`.env` (or its own environment).

### Environment variables

- `DISCORD_TOKEN` — bot token (required)
- `DISCORD_SERVER_ID` — default server ID

At startup, `disc` reads a `.env` file in the current directory (if present)
without overriding variables already set in your shell.

### First-time setup

```bash
disc init
```

Prompts for your bot token, validates it, and writes `DISCORD_TOKEN=...` to
a `.env` file in the current directory. It then prints an invite link to add
the bot to a server.

The `.env` file is git-ignored so your token never gets committed.

## Usage

### Status

Show connection status and the servers the bot belongs to.

```bash
disc status
```

### Invite

Generate an OAuth2 invite link to add the bot to a server.

```bash
disc invite
```

### Channels

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
```

All channel commands accept `--server <server-id>` to override the default server. Mutating commands ask for confirmation unless `--yes` is given.

### Roles

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
```

All role commands accept `--server <server-id>`. Mutating commands ask for confirmation unless `--yes` is given.

### Scheduled events

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

Event types are `voice` (default), `stage`, and `external`. External events require a `--location` and an `--end` time. Use `--status` on update to start, end, or cancel an event (`scheduled`, `active`, `completed`, `canceled`). All event commands accept `--server <server-id>`; mutating commands ask for confirmation unless `--yes` is given.

## Server ID resolution

Commands that operate on a server resolve the server ID in this order:

1. The `--server` flag
2. The `DISCORD_SERVER_ID` environment variable (which may come from `.env`)
3. The single server the bot is a member of, if it's in exactly one server

If the bot is in more than one server and no explicit server ID is given,
disc asks you to either set `DISCORD_SERVER_ID` or pass `--server`.

## Confirmation prompts

Every mutating command (create, update, delete, move) prompts for confirmation before making changes. Pass `-y` / `--yes` to skip the prompt — useful for scripting.

## Config

- Source: `DISCORD_TOKEN` and `DISCORD_SERVER_ID` environment variables
- `.env` file: a local `.env` in the working directory is loaded at startup if present (existing env vars win)
- `disc init` writes `DISCORD_TOKEN` to `.env`
- `.env` is git-ignored

## Related projects

- [nomad-server](https://github.com/metruzanca/nomad-server) — the Discord bot and Terraform server config this tool supports (channel/role discovery for Terraform imports)
