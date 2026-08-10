# disc

A CLI for managing Discord servers via a bot.

Manage Discord servers, channels and roles. Designed for automation and agent-based workflows.

## Features

- **Server discovery** — list channels and roles to see the current structure of a server
- **Channel management** — create, update, delete, and move channels
- **Role management** — create, update, delete, and inspect roles
- **Automation friendly** — non-interactive via `--yes` for scripted use
- **Flexible config** — token and server ID from a config file or environment variables

## Requirements

- [Go](https://go.dev/dl/) 1.25+
- A Discord bot token from the [Discord Developer Portal](https://discord.com/developers/applications)
  - Create an application → open the **Bot** tab → **Reset Token** / **Copy** to get your token
  - Grant the bot the **Manage Channels**, **Manage Roles**, **View Channels**, **Send Messages**, and **Read Message History** permissions when inviting it

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

### First-time setup

```bash
disc init
```

Prompts for your bot token and saves it to `~/.config/disc/disc.toml`. It validates the token and prints an invite link to add the bot to a server.

### Default server

```bash
disc config set-server <server-id>
disc config clear-server
```

### Configuration file

`disc` reads its token and default server **only** from the config file
(`~/.config/disc/disc.toml`). There are no environment variables involved.
This keeps every instance local and self-contained, so you can run multiple
bots with the same `disc` binary without them stepping on each other — each
instance uses its own config file.

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

## Server ID resolution

Commands that operate on a server resolve the server ID in this order:

1. The `--server` flag
2. The `server_id` value in the config file

## Confirmation prompts

Every mutating command (create, update, delete, move) prompts for confirmation before making changes. Pass `-y` / `--yes` to skip the prompt — useful for scripting.

## Config file

- Location: `~/.config/disc/disc.toml`
- Format: TOML with `token` and `server_id` keys
- Permissions: created with `0600` so the token is not world-readable

## Related projects

- [nomad-server](https://github.com/metruzanca/nomad-server) — the Discord bot and Terraform server config this tool supports (channel/role discovery for Terraform imports)
