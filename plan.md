Actually the init local/global is a bit confusing. Especially when the server.json isn't created at this stage. And init/config are confusing. I want to change how the cli works for new users.

So after installation, no init needed.

Users will instead run `disc server new` which will ask the user for a discord token (like init) and store it globally in a new file: `~/.config/disc/bots.json`. This file will hold all the discord tokens for all the bots in an easily accessible location. likely with a structure like this: (using typescript for simplicity, but we're using go of course)
```ts
type BotJson = {
  bots: Record<string, {
    id: string // discord bot user id
    discord_token: string // token
    servers: Record<string, {
      local_path?: string
    }> // servers currently under management, not all servers bot is in
  }>
  active_bot: string
  active_server: string
}
```

Then the command will try to see if it can find the server the bot is added to. If the bot is in multiple, it will ask the server which they want to manage. Once a server is established we'll at this point ask if they want to have the server.json be local or global. If global create `~/.config/disc/servers/<server_id>/server.json`. Otherwise `./server.json` is created and we save `local_path` in the `bot.json`.

Whenever we use commands that change the config, if theres a local `server.json` we can assume the user wants to edit that server. So without changing the active_server/active_bot, we'll use the information from server.json to use the correct bot. If no local server.json, we'll grab our `~/.config/disc/bots.json`, read active bot+server. A user can at any time switch servers with `disc server switch` which will give them a list of servers from the `bot.json` as well as where they are on disk. (for convenience to the user, if they select a server that is locally managed, we'll `cd $local_path` for them).

To reiterate:
- if `./server.json` is present, use that server.
- else use the global active server+bot
- disc server switch can be used to change to any server, as all are always mentioned in bot.json.

When server.json is created, regardless of location we will automatically run what is currently `disc config pull`.

---

We also want to remove the `config` grouping. here's the commands that should be accessible from `disc`
- channel: manage channels via config
- role: manage roles via config
- event: imperative event CRUD/duplicate
- help
- server: new, switch, pull, push
- invite: shows invite link for current active_bot
- status: like before but now that we know all the bots, we can show all the bots and which servers they manage. Also useful to see if keys have been rotated. For the output of this command we should use the bot token to get a list of servers they're in, then use our internal list to filter the servers and only show those being managed by us.

No longer present:
- completion: this we should keep but hide. This is useful for shell integration but end users shouldn't see this. Using `disc completion` should still work as it does, just don't list it in the help screen.
- config: removed but commands still exist via channel/role
- init: removed, we auto init now so user doesn't get weird "you need to init first" errors.
