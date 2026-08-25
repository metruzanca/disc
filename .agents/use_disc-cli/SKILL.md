---
name: disc
description: Use when managing a Discord server through the `disc` CLI (a Go bot). Covers channels, roles, scheduled events, and server management via bots.json and server.json. Use whenever asked to create, update, delete, list, move, or inspect Discord channels/roles/events, or to snapshot and apply server state with `disc server`.
---

# disc

`disc` is a CLI for managing Discord servers via a bot. Roles and channels are
managed **declaratively** via `disc role add/update/delete` and `disc channel add/update/delete/move`
(edit a `server.json` file, then apply with `disc server push`). Scheduled events
are managed **imperatively** via `disc event` commands. All bots and servers are
registered in `~/.config/disc/bots.json`.

Before running anything, confirm the bot is configured:

```bash
disc status        # checks all bots and shows managed servers
disc --help        # every subcommand also has its own --help
```

## Configuration

`disc` stores all bot tokens and server registry in `~/.config/disc/bots.json`.
No environment variables or `.env` files are used.

Run `disc server new` to add a bot and server (prompts for token and config location).

The server ID is resolved in this order: the `--server` flag → a local
`server.json` in the current directory → the active server from `bots.json`.

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

## Server management

```bash
disc server new                    # Add a new bot and server
disc server switch                # Switch the active server (prints cd hint)
disc server switch --server 123  # Non-interactive switch
disc server list                  # List all managed bots and servers
disc server pull                  # Snapshot live server into server.json
disc server push                  # Apply server.json to the live server
disc server push --dry            # Preview the plan
disc server push --delete-missing  # Also delete resources absent from file
```

`disc server new` flags: `--token` (required in agent mode), `--config-location`
(`local` or `global`), `--server-id` (required when bot is in multiple servers).

`disc server push` flags: `--file`, `--delete-missing`, `--yes`, `--dry`, `--agent`.

## Channels

```bash
# Read live server state
disc channel list                       # grouped by category
disc channel show --channel 123456789   # details incl. permission overwrites

# Edit local config (no server changes until push)
disc channel add --name general
disc channel add --name help --category Staff
disc channel add --name voice-chat --type voice
disc channel update --channel 123456789 --topic "New topic"
disc channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
disc channel delete --channel 123456789
disc channel move --channel 111111111 --position 0
disc channel move --channel 111111111 --category Staff --position 3
```

Channel flags: `--file`, `--name`, `--type` (`text`|`voice`|`category`|`news`|`stage`|`forum`),
`--topic`, `--nsfw`, `--category`, `--position`, `--allow`/`--deny`
(permission overwrites, see below).

## Roles (read-only and config editing)

```bash
# Read live server state
disc role list                          # sorted by hierarchy, highest first
disc role show --role 987654321
disc role perm list                     # exact names for --permissions

# Edit local config (no server changes until push)
disc role add --name Moderator --permissions "Kick Members,Ban Members"
disc role add --name VIP --color FF0000 --hoist
disc role update --role 987654321 --name "New Name"
disc role update --role 987654321 --color 00FF00
disc role update --role 987654321 --position 3
disc role delete --role 987654321
```

Role flags: `--file`, `--name`, `--color` (hex, e.g. `FF0000`), `--hoist`,
`--mentionable`, `--permissions` (comma-separated exact names), `--position`.

## Permission overwrites

`--allow` and `--deny` take `Role:Perm,Perm` values and are repeatable. They
replace the given side of a role's overwrite while preserving the other side.
`--no-overwrite <Role>` removes a role's overwrite. Member-level overwrites are
not managed by disc.

```bash
disc channel update --channel 123456789 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
```

To find exact permission names:

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
disc server new --agent --token <token> --config-location local
disc event add --agent --name "Coffee Break" --start "2026-01-15 19:00" --yes
disc server push --agent --dry
disc server push --agent --yes
```

Rules:
- With `--agent` and a missing required parameter, the CLI returns a descriptive
  error naming the missing field rather than launching a form or blocking.
- With `--agent` and a mutating command that would otherwise confirm, pass
  `--yes` to apply or `--dry` to preview; otherwise it errors with instructions.

## Automation tips

- Use `--dry` to preview and `-y` / `--yes` to apply without prompting — never
  rely on the interactive prompt (it blocks on stdin). Add `--agent` to force
  non-interactive behavior unconditionally.
- Always preview with `--dry` and get user sign-off before applying.
- Config edits are local and non-interactive; only `push` touches the server.
- Write to a specific file per environment:

```bash
disc server pull --file prod.json
disc role add --file prod.json --name Helper --permissions "Send Messages"
disc server push --file prod.json --dry
disc server push --agent --file prod.json --delete-missing --yes
```

## Rules of thumb

- Always confirm with `disc status` first if unsure bots are configured.
- Look up permission names with `disc role perm list` rather than guessing.
- For destructive config pushes, always preview with `disc server push --dry`
  first, and only add `--delete-missing` when deletion is intended.
