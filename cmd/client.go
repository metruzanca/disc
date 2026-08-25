package cmd

import (
	"fmt"

	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
)

// resolveServerID returns the server ID to operate on.
//
// Resolution order:
//  1. the --server flag
//  2. the DISCORD_SERVER_ID environment variable (or .env / global env)
//  3. the single server the bot is a member of (if exactly one)
//
// If the bot is in more than one server and no explicit ID is given, an
// error instructs the user to set DISCORD_SERVER_ID or pass --server.
func resolveServerID(flagValue string, client *discord.Client) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	if id := cfg.ServerID; id != "" {
		return id, nil
	}

	guilds := client.Guilds()
	switch len(guilds) {
	case 0:
		return "", fmt.Errorf("no server ID set and the bot is not in any server; pass --server or set DISCORD_SERVER_ID")
	case 1:
		return guilds[0].ID, nil
	default:
		return "", fmt.Errorf("no server ID set and the bot is in %d servers; pass --server or set DISCORD_SERVER_ID", len(guilds))
	}
}

// newClientAndServer creates a connected client and resolves the server ID.
func newClientAndServer(serverFlag string) (*discord.Client, string, error) {
	if serverFlag != "" {
		if err := config.LoadForServer(serverFlag); err != nil {
			return nil, "", err
		}
	}
	client, err := newClient()
	if err != nil {
		return nil, "", err
	}
	serverID, err := resolveServerID(serverFlag, client)
	if err != nil {
		client.Close()
		return nil, "", err
	}
	return client, serverID, nil
}

// newClient creates a connected client from the configured token.
func newClient() (*discord.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	token := cfg.Token
	if token == "" {
		return nil, fmt.Errorf("no bot token configured; run disc init")
	}
	client, err := discord.New(token)
	if err != nil {
		return nil, err
	}
	if err := client.WaitReady(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}
