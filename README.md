# disc

A CLI for managing Discord servers via a bot.

Manage Discord servers, channels, roles, and scheduled events. Designed for automation and agent-based workflows.

disc uses a **declarative** approach for roles and channels: you edit a local `server.json` file with `disc channel add`, `disc role add` etc., then apply it to the server with `disc server push`. All changes are saved to the config file so nothing is lost.

## Features

- **Multi-bot** — manage multiple Discord bots from one CLI, each with their own servers
- **Declarative config** — describe your server in `server.json`, apply with `disc server push`
- **Server discovery** — list channels and roles to see the current structure of a server
- **Channel/role inspection** — read-only `channel list`, `role list`, etc.
- **Scheduled event management** — imperative commands for events (create, update, delete, copy)
- **Automation friendly** — non-interactive via `--yes` for scripted use
- **Flexible config** — local `server.json` (per-project) or global (`~/.config/disc/servers/<id>/server.json`)

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

Install the agent skill (for use with opencode, Claude Code, Codex and other skill-aware agents)
with:

```bash
npx skills add metruzanca/disc
```

## Configuration

disc is configured through `~/.config/disc/bots.json`, which stores all bot
tokens and the servers under management.

### Config file locations

disc supports two config file locations per server:

- **Local directory** (`server.json` in the current directory) — commit it to version control and manage your server as code. Best for teams.
- **disc config directory** (`~/.config/disc/servers/<server-id>/server.json`) — all server configs live in a known location. Best for simple, personal server usage.

Run `disc server new` to choose your preferred location.

## Quick start

```bash
# Add your first bot and server
disc server new

# Check the connection and which servers the bot belongs to
disc status

# See the structure of a server
disc channel list
disc role list

# Generate an invite link to add the bot to a server
disc invite
```

## Server ID resolution

Commands that operate on a server resolve the server ID in this order:

1. The `--server` flag
2. A local `server.json` in the current directory
3. The active server from `bots.json`

## Confirmation prompts & interactive forms

Every mutating command (create, update, delete) asks for a yes/no
confirmation before making changes. In an interactive terminal, the `event add` command shows a full form prefilled from any flags you pass.

Flags control this for scripting and agents:

- `-y` / `--yes` — skip the prompt and apply immediately.
- `--dry` — print what would happen and exit without changing anything.
- `--agent` — force everything non-interactive (disables forms and prompts). If
  a required parameter is missing, the command returns a descriptive error
  naming the missing field instead of blocking.

### Server setup commands

```bash
# Add a new bot and server
disc server new

# List all managed bots and servers
disc server list

# Snapshot the live server into the config file
disc server pull

# Make changes to roles & channels
...

# Preview what would change
disc server push --dry

# Apply the config to the live server
disc server push
```

### Edit the config

Roles and channels are edited via subcommands — nothing changes on the server until you push.

```bash
# Roles
disc role add --name Moderator --permissions "Kick Members,Ban Members"
disc role update --role 123456789 --color 00FF00
disc role delete --role 123456789

# Channels
disc channel add --name help --category Staff
disc channel update --channel 111111111 --topic "Welcome"
disc channel delete --channel 111111111
```

Find exact permission names with:

```bash
disc role perm list
```

### Config file resolution

When no `--file` is given, disc resolves the config file like this:

1. `--file` flag wins.
2. A local `server.json` in the current directory wins (so a committed file in a repo is always used).
3. Otherwise, uses the registered local path for the active server, or `~/.config/disc/servers/<id>/server.json` for globally-managed servers.

### Multiple servers

disc manages multiple servers by storing each server's config and bot token in `bots.json`. Use `disc server switch` to change between them. Each server can be locally or globally managed independently.

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
