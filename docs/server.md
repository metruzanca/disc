# disc server example

`disc server` manages the bot registry (`bots.json`) and the declarative server
config workflow: snapshot with `disc server pull`, apply with `disc server push`.

## bots.json

All bots and servers are registered in `~/.config/disc/bots.json`. Each bot entry
contains the bot's user ID, token, and the servers it manages (with optional
`local_path` for locally-managed servers). Active bot and server are top-level
pointers.

```json
{
  "bots": {
    "123456789": {
      "id": "123456789",
      "discord_token": "...",
      "servers": {
        "987654321": {
          "name": "My Server",
          "local_path": "/home/me/projects/disc"
        }
      }
    }
  },
  "active_bot": "123456789",
  "active_server": "987654321"
}
```

## Adding a bot and server

```bash
disc server new
```

Prompts for a bot token (or pass `--token`), validates it, lists the bot's
servers (or `--server-id` to specify), and asks for config location:

- **Local**: creates `./server.json` and records `local_path` in bots.json
- **Global**: creates `~/.config/disc/servers/<id>/server.json`

On creation, the live server state is automatically pulled into the new
server.json.

## Switching servers

```bash
disc server switch
```

Lists all managed servers (across all bots) with their location. If a local
server is selected, prints a `cd` hint so you can enter that directory.

Non-interactively:

```bash
disc server switch --server 987654321
```

## Listing managed servers

```bash
disc server list
```

Non-interactive, parseable output showing all bots, servers, locations, and
active markers.

## Pull

Snapshot the live server into server.json (roles + channels together):

```bash
disc server pull
disc server pull --file prod.json
```

The config file's `server_id` and `bot` fields are set from the active
connection.

## Push

Apply server.json to the live server. Reconciles by ID (when present) or name.
Idempotent — running again with no edits produces no changes.

```bash
disc server push
```

Dry run and confirmation:

```bash
disc server push --dry
disc server push --delete-missing
disc server push --yes
```

With `--delete-missing`, resources on the server absent from the file are
deleted. **Managed roles, `@everyone`, system channels, and member-level
overwrites are never deleted.**

## Scripting

All config edits are local and non-interactive; only `push` touches the server.
For automation, preview with `--dry` and apply with `-y`:

```bash
disc server pull --file prod.json
disc role add --file prod.json --name Helper --permissions "Send Messages"
disc server push --file prod.json --dry
disc server push --file prod.json --delete-missing --yes
```
