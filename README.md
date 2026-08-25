# disc

A CLI for managing Discord servers via a bot.

Manage Discord servers, channels, roles, and scheduled events. Designed for automation and agent-based workflows.

disc uses a **declarative** approach: you edit a local `server.json` file with `disc config role/channel` commands, then apply it to the server with `disc config push`. All changes are saved to the config file so nothing is lost.

## Features

- **Declarative config** — describe your server in `server.json`, apply with `disc config push`
- **Server discovery** — list channels and roles to see the current structure of a server
- **Channel/role inspection** — read-only `channel list`, `role list`, etc.
- **Scheduled event management** — imperative commands for events (create, update, delete, copy)
- **Automation friendly** — non-interactive via `--yes` for scripted use
- **Flexible config** — token and server ID via environment variables, local `.env`, or disc config directory (`~/.config/disc/servers/<id>/`)

## Requirements

- [Go](https://go.dev/dl/) 1.25+
- A Discord bot token from the [Discord Developer Portal](https://discord.com/developers/applications)
  - Create an application → open the **Bot** tab → **Reset Token** / **Copy** to get your token
  - Grant the bot the **Manage Channels**, **Manage Roles**, **View Channels**, **Send Messages**, **Read Message History**, and **Create Events** permissions when inviting it

## Installation

Prebuilt binaries for Linux, macOS, and Windows are published on the
[Releases](https://github.com/metruzanca/disc/releases) page. You can also
install from source:

```bash
go install github.com/metruzanca/disc@latest
```

## Agent skill

Install the agent skill (for use with opencode and other skill-aware agents)
with:

```bash
npx skills add metruzanca/disc
```

## Configuration

disc is configured through environment variables, optionally loaded from a
local `.env` file, or from the disc config directory.

### Environment variables

- `DISCORD_TOKEN` — bot token (required)
- `DISCORD_SERVER_ID` — default server ID
- `DISC_CONFIG_LOCATION` — override config file location: `local` or `global`

### Config file locations

disc supports two config file locations:

- **Local directory** (`server.json` in the current directory) — commit it to version control and manage your server as code. Best for teams.
- **disc config directory** (`~/.config/disc/servers/<server-id>/server.json`) — all server configs live in a known location. Best for simple, personal server usage.

Run `disc init` to choose your preferred location.

### First-time setup

```bash
disc init
```

Prompts for your bot token, validates it, records your chosen config location, and prints an invite link to add the bot to a server.

## Quick start

```bash
# Check the connection and which servers the bot belongs to
disc status

# See the structure of a server
disc channel list
disc role list

# Generate an invite link to add the bot to another server
disc invite
```

## Server ID resolution

Commands that operate on a server resolve the server ID in this order:

1. The `--server` flag
2. The `DISCORD_SERVER_ID` environment variable
3. A local `server.json` in the current directory (if it exists)
4. The disc config directory's default server (if exactly one is configured)

## Confirmation prompts & interactive forms

Every mutating command (create, update, delete) asks for a yes/no
confirmation before making changes. In an interactive terminal, the `event add` command shows a full form prefilled from any flags you pass.

Flags control this for scripting and agents:

- `-y` / `--yes` — skip the prompt and apply immediately.
- `--dry` — print what would happen and exit without changing anything.
- `--agent` — force everything non-interactive (disables forms and prompts). If
  a required parameter is missing, the command returns a descriptive error
  naming the missing field instead of blocking.

## Config file

The server config is a JSON file containing your server's roles and channels:

```json
{
  "server_id": "...",
  "roles": [...],
  "channels": [...]
}
```

### Pull and push

```bash
# Snapshot the live server into the config file
disc config pull

# Apply the config to the live server
disc config push

# Preview what would change
disc config push --dry

# Apply without prompting
disc config push --yes

# Also delete server resources absent from the config file
disc config push --delete-missing --yes
```

### Edit the config

Roles and channels are edited via subcommands — nothing changes on the server until you push.

```bash
# Roles
disc config role add --name Moderator --permissions "Kick Members,Ban Members"
disc config role update --role 123456789 --color 00FF00
disc config role delete --role 123456789

# Channels
disc config channel add --name help --category Staff
disc config channel update --channel 111111111 --topic "Welcome"
disc config channel delete --channel 111111111
```

Find exact permission names with:

```bash
disc role perm list
```

To find the exact `--role` / `--channel` IDs, use `disc config pull` to snapshot the server, then look them up in `server.json`.

### Config file resolution

When no `--file` is given, disc resolves the config file like this:

1. `--file` flag wins.
2. A local `server.json` in the current directory wins (so a committed file in a repo is always used).
3. Otherwise, if `DISCORD_SERVER_ID` is set or only one disc config directory server exists, uses `~/.config/disc/servers/<id>/server.json`.

You can also set `DISC_CONFIG_LOCATION=global` or `local` to control the default.

### Multiple servers

disc manages multiple servers by storing each server's config and bot token in its own directory under `~/.config/disc/servers/`. Set `DISCORD_SERVER_ID` or pass `--server` to switch between them.

## Scheduled events

Events are managed directly (imperatively) since they have a lifecycle (scheduled → active → completed/canceled):

```bash
disc event list
disc event add --name "Coffee Break" --start "2026-01-15 19:00" --channel 123456789
disc event update --event 987654321 --name "New Name"
disc event delete --event 987654321
```

## Documentation

The CLI is designed with a `--help` screen for every subcommand, so the best
way to understand how to do things is to ask the CLI itself:

```bash
disc --help
disc channel list --help
```

We also publish a full command reference at [docs/commands.md](./docs/commands.md)
for anyone who prefers to read it from the web.
