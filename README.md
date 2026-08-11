# disc

A CLI for managing Discord servers via a bot.

Manage Discord servers, channels, roles, and scheduled events. Designed for automation and agent-based workflows.

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

## Configuration

`disc` is configured entirely through environment variables, optionally
loaded from a local `.env` file in the directory you run it from. 

### Environment variables

- `DISCORD_TOKEN` — bot token (required)
- `DISCORD_SERVER_ID` — default server ID

### First-time setup

```bash
disc init
```

Prompts for your bot token, validates it, and writes `DISCORD_TOKEN=...` to
a `.env` file in the current directory. It then prints an invite link to add
the bot to a server.

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
2. The `DISCORD_SERVER_ID` environment variable (which may come from `.env`)
3. The single server the bot is a member of, if it's in exactly one server

If the bot is in more than one server and no explicit server ID is given,
disc asks you to either set `DISCORD_SERVER_ID` or pass `--server`.

## Confirmation prompts

Every mutating command (create, update, delete, move) prompts for confirmation before making changes. Pass `-y` / `--yes` to skip the prompt — useful for scripting.

## Exporting and importing

Channels and roles can be snapshotted to JSON and reapplied.

```bash
# Export the current structure to JSON (channels.json / roles.json by default)
disc channel export
disc role export

# Import a definition: a dry run is shown first, then type "apply" to proceed
disc channel import
disc role import
```

Import matches channels and roles by name, so it is **idempotent** — running
it again produces no changes, and a partial setup is left untouched. Deleting
resources absent from the file only happens with `--delete-missing`; when it
does, the dry run warns (in red) that deletion is non-reversible before asking
you to type `apply`.

## Documentation

The CLI is designed with a `--help` screen for every subcommand, so the best
way to understand how to do things is to ask the CLI itself:

```bash
disc --help
disc channel add --help
```

We also publish a full command reference at [docs/commands.md](./docs/commands.md)
for anyone who prefers to read it from the web.

