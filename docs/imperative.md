# Imperative example: building a programming community

This example walks through creating a Discord community from scratch with the
imperative CLI. Each command applies immediately to the live server — the same
pattern used when working on an existing server one change at a time.

We will model a community loosely based on _The Programmer's Hangout_: a few
staff roles, a handful of categories, and language channels for JavaScript, Go,
and Rust.

## 1. Create the categories

Categories need to exist before you can place channels in them. Create each
one and note its ID from the output.

```bash
disc channel add --name Info --type category
# → Created category 'Info' (1283019827301283)

disc channel add --name General --type category
# → Created category 'General' (1283019827301284)

disc channel add --name Languages --type category
# → Created category 'Languages' (1283019827301285)

disc channel add --name Voice --type category
# → Created category 'Voice' (1283019827301286)
```

## 2. Create the roles

Build the role hierarchy from least to most privileged so that the Moderator
and Admin roles land above Member in the list.

```bash
disc role add --name Member \
  --permissions "View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links,Connect"

disc role add --name Moderator --color 3498DB --hoist \
  --permissions "Kick Members,Ban Members,Manage Messages,Manage Nicknames,Mute Members,Deafen Members,Move Members"

disc role add --name Admin --color E74C3C --hoist --mentionable \
  --permissions "Administrator"
```

List the roles to get their IDs — you'll need them when setting channel
overwrites.

```bash
disc role list
#  → Admin          e74c3c  hoist  mentionable  ID 1283019827301300
#  → Moderator      3498db  hoist               ID 1283019827301301
#  → Member                                   ID 1283019827301302
#  → @everyone                                 ID 1283019827301303
```

## 3. Lock down the categories with overwrites

By default `@everyone` can see new channels. Set category-level overwrites so
that each role inherits the right defaults for all child channels.

```bash
# Info: @everyone reads, nobody else sends messages
disc channel update --channel 1283019827301283 \
  --allow "@everyone:View Channels,Read Message History" \
  --deny "@everyone:Send Messages"

# General: Members and above can talk, @everyone cannot
disc channel update --channel 1283019827301284 \
  --deny "@everyone:View Channels"

# Languages: same as General
disc channel update --channel 1283019827301285 \
  --deny "@everyone:View Channels"

# Voice: Members can connect, @everyone cannot
disc channel update --channel 1283019827301286 \
  --deny "@everyone:View Channels,Connect"
```

## 4. Create the channels

Now add the text and voice channels under each category. Channels inherit
permission overwrites from their parent category; add role-specific overwrites
where needed.

```bash
# --- Info channels ---
disc channel add --name rules --category 1283019827301283
disc channel add --name announcements --category 1283019827301283 \
  --deny "@everyone:Send Messages,Add Reactions"

# --- General channels ---
disc channel add --name general --category 1283019827301284 \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"
disc channel add --name off-topic --category 1283019827301284 \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"

# --- Language channels ---
disc channel add --name javascript --category 1283019827301285 \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"
disc channel add --name go --category 1283019827301285 \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"
disc channel add --name rust --category 1283019827301285 \
  --allow "Member:View Channels,Send Messages,Read Message History,Add Reactions,Attach Files,Embed Links"

# --- Voice channels ---
disc channel add --name "General Voice" --type voice --category 1283019827301286 \
  --allow "Member:View Channels,Connect,Speak"
```

## 5. Verify the result

```bash
disc channel list
# Info
#   📄 rules
#   📄 announcements
# General
#   📄 general
#   📄 off-topic
# Languages
#   📄 javascript
#   📄 go
#   📄 rust
# Voice
#   🔈 General Voice

disc role list
# Admin       e74c3c  hoist  mentionable
# Moderator   3498db  hoist
# Member
# @everyone
```

## When to use the imperative workflow

The imperative commands are a good fit when you are making small, incremental
changes to an already-live server — renaming a channel, adjusting a role colour,
or adding a single new channel. Because each command hits the API immediately,
you see the result right away and can undo a mistake with an opposite command.

For building an entire community from scratch or maintaining a reproducible
configuration, use the [declarative workflow](./declarative.md) instead.
