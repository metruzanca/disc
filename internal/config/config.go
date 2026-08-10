// Package config handles loading and saving disc's persistent configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultServerIDEnv is the env var for the default server.
	DefaultServerIDEnv = "DISCORD_SERVER_ID"
	// TokenEnv is the env var for the bot token.
	TokenEnv = "DISCORD_TOKEN"
)

// Config is the persisted disc configuration.
type Config struct {
	Token    string `toml:"token"`
	ServerID string `toml:"server_id"`
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		// fall back to the current user home
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "disc", "disc.toml")
}

// Load reads the config file. It returns an empty config if the file
// does not exist.
func Load() (Config, error) {
	var cfg Config
	path := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}
	return cfg, nil
}

// Save writes the config to disk with 0600 permissions, creating the
// parent directory if necessary.
func Save(cfg Config) error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// EffectiveToken returns the bot token, preferring the env var over the config file.
func (c Config) EffectiveToken() string {
	if v := os.Getenv(TokenEnv); v != "" {
		return v
	}
	return c.Token
}

// EffectiveServerID returns the server ID, preferring the env over config.
func (c Config) EffectiveServerID() string {
	if v := os.Getenv(DefaultServerIDEnv); v != "" {
		return v
	}
	return c.ServerID
}
