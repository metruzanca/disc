package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
)

func resolveServerID(serverFlag string) (string, error) {
	if serverFlag != "" {
		return serverFlag, nil
	}

	if _, err := os.Stat("server.json"); err == nil {
		var cfg serverConfigFile
		if err := readJSONFile("server.json", &cfg); err == nil && cfg.Server != "" {
			return cfg.Server, nil
		}
	}

	bf, err := loadBots()
	if err != nil || bf == nil {
		return "", fmt.Errorf("no server configured; run 'disc server new' or use --server")
	}
	if bf.ActiveServer != "" {
		return bf.ActiveServer, nil
	}
	return "", fmt.Errorf("no server configured; run 'disc server new' or use --server")
}

func resolveBotForServer(serverID string) (string, error) {
	bf, err := loadBots()
	if err != nil {
		return "", err
	}
	if bf == nil || bf.Bots == nil || len(bf.Bots) == 0 {
		return "", fmt.Errorf("no bots configured; run 'disc server new'")
	}

	if serverID != "" {
		if botID, ok := bf.BotManagingServer(serverID); ok {
			return botID, nil
		}
	}

	if bf.ActiveBot != "" {
		return bf.ActiveBot, nil
	}

	return "", fmt.Errorf("cannot determine which bot to use; run 'disc server new' or use --server")
}

func newClientAndServer(serverFlag string) (*discord.Client, string, error) {
	serverID, err := resolveServerID(serverFlag)
	if err != nil {
		return nil, "", err
	}

	botID, err := resolveBotForServer(serverID)
	if err != nil {
		return nil, "", err
	}

	bf, _ := loadBots()
	botEntry, _ := bf.GetBot(botID)
	token := botEntry.Token
	if token == "" {
		return nil, "", fmt.Errorf("no token found for bot %s; run 'disc server new'", botID)
	}

	client, err := discord.New(token)
	if err != nil {
		return nil, "", fmt.Errorf("invalid token: %w", err)
	}
	if err := client.WaitReady(); err != nil {
		client.Close()
		return nil, "", err
	}
	return client, serverID, nil
}

func resolveActiveClient() (*discord.Client, error) {
	bf, err := loadBots()
	if err != nil {
		return nil, err
	}
	if bf == nil || bf.Bots == nil || len(bf.Bots) == 0 {
		return nil, fmt.Errorf("no bots configured; run 'disc server new'")
	}

	botID := bf.ActiveBot
	if botID == "" {
		return nil, fmt.Errorf("no active bot; run 'disc server new' or 'disc server switch'")
	}

	botEntry, ok := bf.GetBot(botID)
	if !ok {
		return nil, fmt.Errorf("bot %s not found in bots.json", botID)
	}

	client, err := discord.New(botEntry.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if err := client.WaitReady(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func resolveActiveServerID() (string, error) {
	bf, err := loadBots()
	if err != nil {
		return "", err
	}
	if bf == nil || bf.ActiveServer == "" {
		return "", fmt.Errorf("no active server; run 'disc server switch'")
	}
	return bf.ActiveServer, nil
}

func resolveConfigFileForServer(serverFlag, fileFlag string) (string, string, error) {
	serverID, err := resolveServerID(serverFlag)
	if err != nil {
		return "", "", err
	}
	botID, err := resolveBotForServer(serverID)
	if err != nil {
		return "", "", err
	}

	bf, _ := loadBots()

	if fileFlag != "" {
		return fileFlag, botID, nil
	}

	if _, err := os.Stat("server.json"); err == nil {
		return "server.json", botID, nil
	}

	if bf != nil {
		if entry, ok := bf.GetServerEntry(botID, serverID); ok && entry.LocalPath != "" {
			return filepath.Join(entry.LocalPath, "server.json"), botID, nil
		}
		if p, err := config.ServerConfigFile(serverID); err == nil {
			return p, botID, nil
		}
	}

	return "", botID, fmt.Errorf("cannot resolve config file path")
}
