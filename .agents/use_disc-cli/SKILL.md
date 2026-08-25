---
name: disc
description: Use when managing a Discord server through the `disc` CLI (a Go bot). Covers channels, roles, scheduled events, and declarative server config via server.json. Use whenever asked to create, update, delete, list, move, or inspect Discord channels/roles/events, or to snapshot and apply server state with `disc config`.
---

# disc

`disc` is a CLI for managing Discord servers via a bot. Roles and channels are
managed **declaratively** via `disc config` commands (edit a `server.json` file,
then apply with `disc config push`). Scheduled events are managed
**imperatively** via `disc event` commands.

Before running anything, confirm the bot is configured:

```bash
disc status        # checks the connection and lists the bot's servers
disc --help        # every subcommand also has its own --help
```

## Configuration

`disc` is configured through environment variables, optionally loaded from a
local `.env` file, or from the disc config directory (`~/.config/disc/servers/<id>/`).

- `DISCORD_TOKEN` — bot token (required)
- `DISCORD_SERVER_ID` — default server ID
- `DISC_CONFIG_LOCATION` — `local` (current dir `server.json`) or `global` (disc config dir)

Set up from scratch with `disc init` (prompts for token and config location).

The server ID is resolved in this order: the `--server` flag → `DISCORD_SERVER_ID`
→ a local `server.json` in the current directory → the single server in the disc
config directory. If the bot is in more than one server and no ID is given,
an explicit `--server` or env var is required.

## Confirmation prompts & interactive forms

Humans run mutating commands with no flags and get an interactive experience:
a `y/N` confirmation, and a full prefilled form for `event add`. Agents should use
the flag-based interface so nothing blocks on stdin:

- `--yes` / `-y` — skip the prompt and apply immediately.
- `--dry` — print what would happen and exit without prompting or changing
  anything. If `--dry` and `--yes` are both given, `--dry` wins.
- `--agent` — force everything non-interactive. Forms and prompts are disabled;
  if a required parameter is missing, the command returns a descriptive error
  naming the missing field instead of blocking.

Agent workflow for any mutating command: run it with `--dry` to preview, ask
the user to confirm, then run it again with `--yes` to apply.

## Channels (read-only)

```bash
disc channel list                       # grouped by category
disc channel show --channel 123456789   # details incl. permission overwrites
```

## Roles (read-only)

```bash
disc role list                          # sorted by hierarchy, highest first
disc role show --role 987654321
disc role perm list                     # exact names for --permissions
```

## Declarative config

Manage roles and channels as JSON in a single `server.json` (commit this file
and reapply it for reproducible server state). Use `disc config role` and
`disc config channel` subcommands to edit the file; nothing changes on the
server until you run `disc config push`.

```bash
# Snapshot the live server (roles + channels together)
disc config pull
disc config pull --file server.json --server 123456789

# Edit the file (nothing changes on the server until push)
disc config role add --name Helper --permissions "Send Messages,Add Reactions"
disc config role update --role 987654321 --color 00FF00
disc config role update --role 987654321 --position 3
disc config role delete --role 987654321
disc config channel add --name help --category Staff
disc config channel update --channel 123456789 --topic "New topic"
disc config channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
disc config channel update --channel 123456789 --no-overwrite Moderator
disc config channel move --channel 111111111 --category Staff --position 3
disc config channel delete --channel 123456789

# Apply to the server (prints a dry-run plan, then confirms)
disc config push
# Non-interactive: preview with --dry, apply with --yes
disc config push --dry
disc config push --yes
disc config push --agent --file server.json --delete-missing --yes
```

All `disc config role/channel` commands accept `--server <id>` to target a
specific server and `--file <path>` to use a specific config file.

Config flags: `--file` (defaults to `server.json`), `--server`, and on push
`--delete-missing` (non-reversible), `--yes`, `--dry`, and `--agent`.

### How config push works

- **Idempotent** — pushing with no edits produces no changes.
- Entries are matched by ID when present, otherwise by name, so renames update
  in place and a partial setup is left untouched.
- Roles are reordered to match their configured positions.
- Channel permission overwrites are reconciled, resolving roles by name
  (including roles created during the same push).
- `--delete-missing` deletes server resources absent from the file and removes
  live role overwrites absent from the config. It warns in red in the plan.
  **Managed (bot/integration) roles, `@everyone`, system
  channels, and member-level overwrites are never deleted.**
- Entries added with `add` have no server ID until they are pushed and pulled
  again, so they can only be edited by ID after that round trip.

### Agent workflow for applying config

Always preview before changing the server, and get user sign-off:

1. Run `disc config push --dry` to review the planned create/update/delete plan.
2. Ask the user to confirm.
3. On approval, run `disc config push --yes` (add `--delete-missing` only when
   deletion is intended).

## Permission overwrites

`--allow` and `--deny` take `Role:Perm,Perm` values and are repeatable. They
replace the given side of a role's overwrite while preserving the other side.
`--no-overwrite <Role>` removes a role's overwrite. Member-level overwrites are
not managed by disc.

```bash
disc config channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
disc config channel update --channel 123456789 --no-overwrite Moderator
```

To find exact permission names for `--permissions` / `--allow` / `--deny`:

```bash
disc role perm list
```

## Scheduled events (imperative)

```bash
disc event list                         # add --active for scheduled/active only
disc event show --event 987654321

# Voice event hosted in a channel (default type)
disc event add --name "Coffee Break" --start "2026-01-15 19:00" --channel 123456789

# External event (requires --location and --end)
disc event add --name "Meetup" --type external --location "Local Cafe" \
  --start "2026-01-15 19:00" --end "2026-01-15 21:00"

disc event update --event 987654321 --name "New Name"
disc event update --event 987654321 --status active
disc event delete --event 987654321

# Copy an event, overriding any property with the add flags
disc event copy --event 987654321 --name "Coffee Break (Week 2)" --start "2026-01-22 19:00"
```

Event flags: `--name`, `--description`, `--start`, `--end`, `--channel`,
`--type` (`voice`|`stage`|`external`), `--location`, `--status`.

Event `--status` lifecycle: `scheduled`, `active`, `completed`, `canceled`.
`external` events require `--location` and `--end`.

## Agent mode

Pass `--agent` on any command to guarantee zero interactivity for the whole
invocation — even when the CLI thinks stdin is a terminal. Always pass `--agent`
(and use `--yes`/`--dry`) when running as an agent or in a script:

```bash
disc event add --agent --name "Coffee Break" --start "2026-01-15 19:00" --yes
disc config push --agent --dry
disc config push --agent --yes
```

Rules:
- With `--agent` and a missing required parameter, the CLI returns a descriptive
  error naming the missing field rather than launching a form or blocking.
- With `--agent` and a mutating command that would otherwise confirm, pass
  `--yes` to apply or `--dry` to preview; otherwise it errors with instructions.
- `--agent` also makes `disc init` require `--token <token>` instead of reading
  stdin.

## Automation tips

- Use `--dry` to preview and `-y` / `--yes` to apply without prompting — never
  rely on the interactive prompt (it blocks on stdin). Add `--agent` to force
  non-interactive behavior unconditionally.
- Always preview with `--dry` and get user sign-off before applying.
- Config edits are local and non-interactive; only `push` touches the server.
- Write to a specific file per environment:

```bash
disc config pull --file prod.json
disc config role add --file prod.json --name Helper --permissions "Send Messages"
disc config push --file prod.json --dry
disc config push --agent --file prod.json --delete-missing --yes
```

## Rules of thumb

- Always confirm with `disc status` first if unsure the bot is connected.
- Look up permission names with `disc role perm list` rather than guessing.
- For destructive config pushes, always preview with `disc config push --dry`
  first, and only add `--delete-missing` when deletion is intended.
