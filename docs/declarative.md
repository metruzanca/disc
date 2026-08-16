# Declarative example: building a programming community

This example builds the same community as the
[imperative walkthrough](./imperative.md) but uses the declarative `disc config`
workflow. All edits are local — nothing touches the live server until
`disc config push`.

## 1. Snapshot the current server

Start by pulling the existing state. Even a fresh server has `@everyone` and a
default channel.

```bash
disc config pull
# → Wrote server.json (1 role, 1 channel)
```

`server.json` now lives in your working directory. It can be committed to git,
edited by hand, or managed with the config subcommands below.

## 2. Add the roles

Roles and channels are added with `disc config role add` and
`disc config channel add`. These only edit the local file.

```bash
disc config role add --name Member \
  --permissions "View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links,Connect"

disc config role add --name Moderator --color 3498DB --hoist \
  --permissions "Kick Members,Ban Members,Manage Messages,Manage Nicknames,Mute Members,Deafen Members,Move Members"

disc config role add --name Admin --color E74C3C --hoist --mentionable \
  --permissions "Administrator"
```

Unlike the imperative workflow, you don't need to track role IDs. Roles without
an ID are new — they will be created on push and assigned IDs by Discord.

## 3. Add the categories

Categories are regular channels with `--type category`. Reference them by name
when placing child channels.

```bash
disc config channel add --name Info --type category
disc config channel add --name General --type category
disc config channel add --name Languages --type category
disc config channel add --name Voice --type category
```

## 4. Add the channels

Channels are placed under a category with `--category <name>`. Overwrites use
the same `Role:Perm,Perm` syntax as the imperative commands.

```bash
# --- Info channels ---
disc config channel add --name rules --category Info \
  --deny "@everyone:Send Messages"

disc config channel add --name announcements --category Info \
  --deny "@everyone:Send Messages,Add Reactions"

# --- General channels ---
disc config channel add --name general --category General \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"

disc config channel add --name off-topic --category General \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"

# --- Language channels ---
disc config channel add --name javascript --category Languages \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"

disc config channel add --name go --category Languages \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"

disc config channel add --name rust --category Languages \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"

# --- Voice channels ---
disc config channel add --name "General Voice" --type voice --category Voice \
  --allow "Member:View Channels,Connect,Speak"
```

## 5. Set category-level overwrites

Update the categories to deny `@everyone` from viewing channels outside of
`Info`. Child channels inherit these overwrites.

```bash
disc config channel update --channel Info \
  --deny "@everyone:Send Messages"

disc config channel update --channel General \
  --deny "@everyone:View Channels"

disc config channel update --channel Languages \
  --deny "@everyone:View Channels"

disc config channel update --channel Voice \
  --deny "@everyone:View Channels,Connect"
```

`--channel` matches by the channel name when there is no ID yet, so you can
update newly-added entries before they are pushed.

## 6. The resulting config

`server.json` now looks like this:

```json
{
  "server_id": "1049619907212333100",
  "roles": [
    {
      "id": "1049619907212333100",
      "name": "@everyone",
      "permissions": ["View Channels", "Send Messages"]
    },
    {
      "name": "Member",
      "permissions": [
        "View Channels", "Send Messages", "Read Message History",
        "Add Reactions", "Attach Files", "Embed Links", "Connect"
      ]
    },
    {
      "name": "Moderator",
      "color": "3498DB",
      "hoist": true,
      "permissions": [
        "Kick Members", "Ban Members", "Manage Messages",
        "Manage Nicknames", "Mute Members", "Deafen Members", "Move Members"
      ]
    },
    {
      "name": "Admin",
      "color": "E74C3C",
      "hoist": true,
      "mentionable": true,
      "permissions": ["Administrator"]
    }
  ],
  "channels": [
    {
      "id": "1049619907212333101",
      "name": "general",
      "type": "text"
    },
    {
      "name": "Info",
      "type": "category",
      "overwrites": [
        { "role": "@everyone", "deny": ["Send Messages"] }
      ]
    },
    {
      "name": "General",
      "type": "category",
      "overwrites": [
        { "role": "@everyone", "deny": ["View Channels"] }
      ]
    },
    {
      "name": "Languages",
      "type": "category",
      "overwrites": [
        { "role": "@everyone", "deny": ["View Channels"] }
      ]
    },
    {
      "name": "Voice",
      "type": "category",
      "overwrites": [
        { "role": "@everyone", "deny": ["View Channels", "Connect"] }
      ]
    },
    {
      "name": "rules",
      "type": "text",
      "parent": "Info",
      "overwrites": [
        { "role": "@everyone", "deny": ["Send Messages"] }
      ]
    },
    {
      "name": "announcements",
      "type": "text",
      "parent": "Info",
      "overwrites": [
        { "role": "@everyone", "deny": ["Send Messages", "Add Reactions"] }
      ]
    },
    {
      "name": "general",
      "type": "text",
      "parent": "General",
      "overwrites": [
        {
          "role": "Member",
          "allow": [
            "View Channels", "Send Messages", "Read Message History",
            "Add Reactions", "Attach Files", "Embed Links"
          ]
        }
      ]
    },
    {
      "name": "off-topic",
      "type": "text",
      "parent": "General",
      "overwrites": [
        {
          "role": "Member",
          "allow": [
            "View Channels", "Send Messages", "Read Message History",
            "Add Reactions", "Attach Files", "Embed Links"
          ]
        }
      ]
    },
    {
      "name": "javascript",
      "type": "text",
      "parent": "Languages",
      "overwrites": [
        {
          "role": "Member",
          "allow": [
            "View Channels", "Send Messages", "Read Message History",
            "Add Reactions", "Attach Files", "Embed Links"
          ]
        }
      ]
    },
    {
      "name": "go",
      "type": "text",
      "parent": "Languages",
      "overwrites": [
        {
          "role": "Member",
          "allow": [
            "View Channels", "Send Messages", "Read Message History",
            "Add Reactions", "Attach Files", "Embed Links"
          ]
        }
      ]
    },
    {
      "name": "rust",
      "type": "text",
      "parent": "Languages",
      "overwrites": [
        {
          "role": "Member",
          "allow": [
            "View Channels", "Send Messages", "Read Message History",
            "Add Reactions", "Attach Files", "Embed Links"
          ]
        }
      ]
    },
    {
      "name": "General Voice",
      "type": "voice",
      "parent": "Voice",
      "overwrites": [
        {
          "role": "Member",
          "allow": ["View Channels", "Connect", "Speak"]
        }
      ]
    }
  ]
}
```

The file is plain JSON — you can edit it by hand, diff it in pull requests, or
store it alongside your project code.

## 7. Apply to the server

Push reconciles the file with the live server. It prints a dry run showing every
planned change, then asks for confirmation.

```bash
disc config push
```

```text
Dry run for server.json:

  create  role 'Member'
  create  role 'Moderator'
  create  role 'Admin'
  create  channel 'Info'
  create  channel 'General'
  create  channel 'Languages'
  create  channel 'Voice'
  create  channel 'rules'
  create  channel 'announcements'
  create  channel 'general'
  create  channel 'off-topic'
  create  channel 'javascript'
  create  channel 'go'
  create  channel 'rust'
  create  channel 'General Voice'

  Summary: 14 to create, 0 to update, 0 to delete

Apply these changes? [y/N]
```

Confirm and the entire community is created in one pass. Running push again
produces no changes — the config is now the source of truth.

```bash
disc config push
# → Dry run for server.json:
# →   Summary: 0 to create, 0 to update, 0 to delete
```

To add a new language channel later, edit the file and push again. To keep the
file in sync with live changes made outside of disc, pull periodically.

## When to use the declarative workflow

The declarative workflow shines when you want to version your server
configuration, reproduce it across multiple servers, or build an entire
community from a single file. Because all edits are local, you can iterate on
the config without touching Discord, preview the plan with `--dry`, and apply
it all at once.

For quick, one-off changes to a live server, use the
[imperative workflow](./imperative.md) instead.
