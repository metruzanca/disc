package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const botsFileName = "bots.json"

type BotsFile struct {
	Bots         map[string]BotEntry `json:"bots"`
	ActiveBot    string              `json:"active_bot"`
	ActiveServer string              `json:"active_server"`
}

type BotEntry struct {
	ID      string                 `json:"id"`
	Token   string                 `json:"discord_token"`
	Servers map[string]ServerEntry `json:"servers"`
}

type ServerEntry struct {
	Name      string `json:"name"`
	LocalPath string `json:"local_path,omitempty"`
}

func BotsPath() (string, error) {
	gd, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(gd, botsFileName), nil
}

func LoadBots() (*BotsFile, error) {
	path, err := BotsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &BotsFile{Bots: map[string]BotEntry{}}, nil
		}
		return nil, fmt.Errorf("failed to read bots.json: %w", err)
	}
	var bf BotsFile
	if err := json.Unmarshal(data, &bf); err != nil {
		return nil, fmt.Errorf("failed to parse bots.json: %w", err)
	}
	return &bf, nil
}

func SaveBots(bf *BotsFile) error {
	path, err := BotsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(bf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode bots.json: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("failed to write bots.json: %w", err)
	}
	return nil
}

func (bf *BotsFile) UpsertBot(id, token string) {
	if bf.Bots == nil {
		bf.Bots = map[string]BotEntry{}
	}
	entry, exists := bf.Bots[id]
	if !exists {
		entry = BotEntry{ID: id, Servers: map[string]ServerEntry{}}
	}
	entry.ID = id
	entry.Token = token
	bf.Bots[id] = entry
}

func (bf *BotsFile) AddServerToBot(botID, serverID, serverName, localPath string) {
	if bf.Bots == nil {
		bf.Bots = map[string]BotEntry{}
	}
	entry, exists := bf.Bots[botID]
	if !exists {
		entry = BotEntry{ID: botID, Servers: map[string]ServerEntry{}}
	}
	if entry.Servers == nil {
		entry.Servers = map[string]ServerEntry{}
	}
	entry.Servers[serverID] = ServerEntry{Name: serverName, LocalPath: localPath}
	bf.Bots[botID] = entry
}

func (bf *BotsFile) SetActive(botID, serverID string) {
	bf.ActiveBot = botID
	bf.ActiveServer = serverID
}

func (bf *BotsFile) GetBot(id string) (BotEntry, bool) {
	b, ok := bf.Bots[id]
	return b, ok
}

func (bf *BotsFile) GetToken(botID string) (string, bool) {
	b, ok := bf.Bots[botID]
	return b.Token, ok
}

func (bf *BotsFile) GetServerEntry(botID, serverID string) (ServerEntry, bool) {
	b, ok := bf.Bots[botID]
	if !ok {
		return ServerEntry{}, false
	}
	s, ok := b.Servers[serverID]
	return s, ok
}

func (bf *BotsFile) BotManagingServer(serverID string) (string, bool) {
	for id, b := range bf.Bots {
		if _, ok := b.Servers[serverID]; ok {
			return id, true
		}
	}
	return "", false
}

func (bf *BotsFile) AllServers() []struct {
	BotID  string
	Bot    BotEntry
	Server string
	Entry  ServerEntry
} {
	var out []struct {
		BotID  string
		Bot    BotEntry
		Server string
		Entry  ServerEntry
	}
	for botID, bot := range bf.Bots {
		for serverID, entry := range bot.Servers {
			out = append(out, struct {
				BotID  string
				Bot    BotEntry
				Server string
				Entry  ServerEntry
			}{BotID: botID, Bot: bot, Server: serverID, Entry: entry})
		}
	}
	return out
}
