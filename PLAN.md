# PLAN: Reimplement `disc`

## Objective
Reimplement `disc` (github.com/metruzanca/disc) from scratch as a self-owned, reproducible CLI. Recover the exact observable behavior from the compiled binary + config, then rebuild it in `/home/szanca/dev/disc`.

## Tool identity (recovered from binary)
- Module path: `github.com/metruzanca/disc`
- Go 1.25, built with `discordgo`, `cobra`/`pflag`, `BurntSushi/toml` (config), `fatih/color` (output), `gorilla/websocket` (via discordgo).
- Config file: `~/.config/disc/disc.toml` (permissions `0600`) with fields `token`, `server_id`.
- Env vars: `DISCORD_TOKEN`, `DISCORD_SERVER_ID`.

## Project structure
```
disc/
├── go.mod                 module github.com/metruzanca/disc
├── main.go
├── cmd/
│   ├── root.go            cobra root
│   ├── init.go
│   ├── status.go
│   ├── invite.go
│   ├── config.go          set-server / clear-server
│   ├── channel.go         list / add / update / delete / move
│   └── role.go            list / show / add / update / delete
├── internal/
│   ├── config/config.go   load/save ~/.config/disc/disc.toml
│   ├── discord/client.go  discordgo session wrapper
│   └── util/prompt.go     confirmation + color helpers
└── PLAN.md
```

## Dependencies
```
require:
  github.com/bwmarrin/discordgo v0.29.0
  github.com/spf13/cobra
  github.com/spf13/pflag
  github.com/BurntSushi/toml v1.6.0
  github.com/fatih/color
  golang.org/x/crypto           (discordgo dep)
```

## Command surface to implement

### `disc init`
- Prompt for bot token, save to `~/.config/disc/disc.toml` (`token`), print OAuth2 invite link with `Manage Channels` + `Manage Roles` permissions. Accept token via stdin.

### `disc status`
- If token missing → "Status: Not configured, Run disc init". Else connect via websocket, print "Status: Connected", bot name, and list of servers (id + name) from `State.Guilds`.

### `disc invite`
- Print OAuth2 invite link with channel/role management permissions.

### `disc config set-server <id>` / `disc config clear-server`
- Persist/clear `server_id` in config.

### `disc channel list`
- Group channels by category (reproduce "No Category" grouping). For each: name, ID. (Verify: only list channels; the current binary showed text+voice and category labels.)

### `disc channel add`
- Flags: `--server`, `--name` (req), `--type` text|voice (default text), `--category`, `-y`. Confirmation prompt unless `-y`.

### `disc channel update`
- Flags: `--channel` (req), `--name`, `--topic`, `--category`, `--nsfw`, `-y`. Confirmation prompt.

### `disc channel delete`
- Flags: `--channel` (req), `-y`. Confirmation prompt.

### `disc channel move`
- Flags: `--channel` (req), `--position` (default -1), `--category` (use `none` to detach), `--server`, `-y`.

### `disc role list`
- Sort by hierarchy (highest first). For each: name, ID, and `[managed]` marker for managed roles (matches observed output).

### `disc role show`
- Flags: `--role` (req), `--server`. Print name, ID, position, color, hoisted, mentionable, managed, permissions (with named flags like `+ Administrator`).

### `disc role add` / `disc role update` / `disc role delete`
- Flags: `--name`, `--color` (hex), `--hoist`, `--mentionable`, `--role`, `--server`, `-y`. Confirmation prompt.

## Cross-cutting behavior
- **Server resolution order:** `--server` flag → `DISCORD_SERVER_ID` env → `server_id` from config.
- **Token resolution order:** `DISCORD_TOKEN` env → config file token.
- **Confirmation prompts** on all mutating commands unless `-y/--yes`.
- **Output style:** colorized via `fatih/color`; use "guild" internally (matches Discord API).
- Config writes use `0600` perms; directory `~/.config/disc/`.

## Verification
1. `go build ./...` and `go vet ./...` pass.
2. `go build -o disc . && ./disc status` → matches "Connected / Nomad Coworking Group".
3. `./disc channel list`, `./disc role list` output matches the recorded real output.
4. Sanity-check non-mutating commands against the live server (bot is already added).

## Non-goals (for now)
- No mutating actions against live servers during dev (only dry-run/confirm). Confirm against real server only at user's approval.
- Match output formatting exactly is best-effort from binary; acceptable to approximate colors/labels.

## Decisions
- Config persists as TOML at `~/.config/disc/disc.toml`.
- Commands are functionally identical to the original; output fidelity is best-effort.
- No extra `--json` flag (keep minimal to match original).
