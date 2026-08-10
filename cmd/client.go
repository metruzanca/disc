package cmd

import (
	"fmt"

	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
)

// resolveServerID returns the server ID, preferring the flag value
// over the configured default.
func resolveServerID(flagValue string) (string, error) {
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
	return "", fmt.Errorf("no server ID set; pass --server or run disc config set-server")
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
