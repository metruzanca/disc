# disc

A CLI for managing Discord servers via a bot.

Manage Discord servers, channels, roles, and scheduled events. Designed for automation and agent-based workflows.

The CLI supports two ways to work with the server:

- **Imperative** — run a command and it changes the server immediately (e.g.
  `disc channel add`, `disc role update`).
- **Declarative** — describe the desired state in a `server.json` file with
  `disc config`, then apply it with `disc config push`.

## Features

- **Two usage modes** — imperative commands for one-off changes, or declarative `disc config` for reproducible, versionable server state
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

Prebuilt binaries for Linux, macOS, and Windows are published on the
[Releases](https://github.com/metruzanca/disc/releases) page. You can also
install from source:

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

## Confirmation prompts & interactive forms

Every mutating command (create, update, delete, move) asks for a yes/no
confirmation before making changes. In an interactive terminal, the `add`
commands (`channel add`, `role add`, `event add`) instead show a full form
prefilled from any flags you pass — with multi-select pickers for role/channel
permissions.

Flags control this for scripting and agents:

- `-y` / `--yes` — skip the prompt and apply immediately.
- `--dry` — print what would happen and exit without changing anything.
- `--agent` — force everything non-interactive (disables forms and prompts). If
  a required parameter is missing, the command returns a descriptive error
  naming the missing field instead of blocking on stdin.

Pass both `--dry` and `--yes` and `--dry` wins, so `disc ... --dry --yes` just
shows the plan. Humans omit the flags and get the interactive prompt/form;
agents and scripts use `--agent` (optionally with `--yes` to apply, `--dry` to
preview) to guarantee the CLI never blocks.

## Declarative config

Roles and channels can be managed declaratively through a single JSON file
(`server.json`), combining what used to be separate snapshots into one
source of truth that can be committed and reapplied.

```bash
# Snapshot the live server into server.json (roles + channels together)
disc config pull

# Make declarative edits to the config file instead of the server
disc config role add --name Moderator --permissions "Kick Members,Ban Members"
disc config role update --role 123456789 --color 00FF00
disc config role update --role 123456789 --position 3
disc config role delete --role 123456789
disc config channel add --name help --category Staff
disc config channel move --channel 111111111 --position 0
disc config channel update --channel 111111111 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
disc config channel update --channel 111111111 --no-overwrite Moderator
disc config channel delete --channel 111111111

# Apply the config to the live server (shows a dry run first, then confirms)
disc config push
disc config push --delete-missing
# Non-interactive: preview with --dry, apply with --yes
disc config push --dry
disc config push --yes
disc config push --delete-missing --yes
```

`disc config pull` includes each role and channel's server ID, role/channel
positions, and channel permission overwrites, so edits can reference resources
by ID. `disc config push` matches entries by ID when present and by name
otherwise, so it is **idempotent** — running it again produces no changes, and
a partial setup is left untouched. It prints a dry-run plan and confirms before
applying (or `--yes` skips the prompt, `--dry` shows the plan only). Deleting
resources absent from the file only
happens with `--delete-missing`; when it does, the dry run warns (in red) that
deletion is non-reversible. Managed roles,
`@everyone`, system channels, and member-level permission overwrites are never
deleted.

Channel permission overwrites are managed per role. On `push`, overwrites
absent from the config are preserved unless `--delete-missing` is given.

To find the exact permission names used by `--permissions` (and `--allow` /
`--deny`), run:

```bash
disc role perm list
```

Imperatively, channel permission overwrites can be set at creation or update,
and channel details inspected, like so:

```bash
disc channel add --name general --allow "Moderator:Send Messages"
disc channel update --channel 111111111 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
disc channel show --channel 111111111
```

## Documentation

The CLI is designed with a `--help` screen for every subcommand, so the best
way to understand how to do things is to ask the CLI itself:

```bash
disc --help
disc channel add --help
```

We also publish a full command reference at [docs/commands.md](./docs/commands.md)
for anyone who prefers to read it from the web.

