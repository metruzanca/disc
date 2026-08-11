# disc config example

`disc config` manages the server's roles and channels as declarative JSON in a
single file (`server.json` by default). This is an end-to-end walkthrough of
the workflow: pull the live state, make declarative edits, then push.

## 1. Snapshot the server

Start by pulling the current server state into `server.json`.

```bash
disc config pull
```

`server.json` combines roles and channels, each carrying its server ID so the
edit commands can reference resources the same way the server CLI does:

```json
{
  "roles": [
    {
      "id": "1049619907212333100",
      "name": "@everyone",
      "permissions": ["View Channels", "Send Messages"]
    },
    {
      "id": "1049620503430934540",
      "name": "Moderator",
      "color": "FF0000",
      "hoist": true,
      "mentionable": false,
      "position": 5,
      "permissions": ["Kick Members", "Ban Members", "Manage Messages"]
    }
  ],
  "channels": [
    {
      "id": "1049619907212333101",
      "name": "Staff",
      "type": "category"
    },
    {
      "id": "1049619907212333102",
      "name": "general",
      "type": "text",
      "topic": "General discussion",
      "nsfw": false,
      "parent": "Staff",
      "position": 0,
      "overwrites": [
        { "role": "Moderator", "allow": ["Send Messages", "Manage Messages"], "deny": ["Attach Files"] },
        { "role": "@everyone", "allow": ["View Channels"] }
      ]
    }
  ]
}
```

Managed (integration/bot) roles, unmanaged channel kinds, and member-level
permission overwrites are skipped.

## 2. Find permission names

Roles' `--permissions` takes exact permission names. List them to look one up:

```bash
disc role perm list
```

```text
  Administrator
  Manage Server
  View Audit Log
  Manage Roles
  ...
```

## 3. Edit the config declaratively

These commands only edit `server.json`; nothing changes on the server until
`disc config push` runs.

### Roles

```bash
# Add a new role (created on the server at push time; has no ID yet)
disc config role add --name Helper --permissions "Send Messages,Add Reactions"

# Update an existing role by its ID
disc config role update --role 1049620503430934540 --color 00FF00
disc config role update --role 1049620503430934540 --hoist=false
disc config role update --role 1049620503430934540 --permissions "Kick Members"
disc config role update --role 1049620503430934540 --position 3

# Remove a role
disc config role delete --role 1049620503430934540
```

### Channels

```bash
# Add a channel under a category (referenced by name)
disc config channel add --name help --category Staff
disc config channel add --name voice-chat --type voice

# Update a channel by its ID
disc config channel update --channel 1049619907212333102 --topic "New topic"
disc config channel update --channel 1049619907212333102 --nsfw
disc config channel update --channel 1049619907212333102 --category Staff

# Set permission overwrites (role-based)
disc config channel update --channel 1049619907212333102 --allow "Moderator:Send Messages" --deny "@everyone:Attach Files"
disc config channel update --channel 1049619907212333102 --no-overwrite Moderator

# Move a channel to a position or category
disc config channel move --channel 1049619907212333102 --position 2
disc config channel move --channel 1049619907212333102 --category Staff --position 3

# Remove a channel
disc config channel delete --channel 1049619907212333102
```

`--allow` and `--deny` take `Role:Perm,Perm` values and are repeatable; they
replace the given side of a role's overwrite while preserving the other.
`--no-overwrite` removes a role's overwrite. Member-level overwrites are not
managed.

> New entries added with `add` have no server ID until they are pushed and
> pulled again, so they can only be edited by ID after that round trip.

## 4. Apply the config to the server

Push reconciles `server.json` with the live server. A dry run shows every
planned create/update/delete; type `apply` to proceed.

```bash
disc config push
```

```text
Dry run for /path/to/server.json:

  create  role 'Helper'
  update  role 'Moderator'
  update  channel 'general'

  Summary: 1 to create, 2 to update, 0 to delete

Type 'apply' to continue, or anything else to cancel:
```

Push is **idempotent**: running it again with no edits produces no changes.
It matches each entry by ID when present and by name otherwise, so a rename
updates the role in place. Roles are reordered to match their configured
positions, and channel permission overwrites are reconciled (resolving roles by
name, including roles created during the same push).

To also delete server resources that are absent from the file, pass
`--delete-missing`. This is non-reversible, and the dry run warns in red. It
also removes live role permission overwrites absent from the config. Managed
roles, `@everyone`, system channels, and member-level overwrites are never
deleted.

```bash
disc config push --delete-missing
```

## Scripting

All config edits are local and non-interactive. Push is the only command that
touches the server, and `-y` skips its confirmation for automation:

```bash
disc config pull --file prod.json
disc config role add --file prod.json --name Helper --permissions "Send Messages"
disc config push --file prod.json --delete-missing --yes
```
