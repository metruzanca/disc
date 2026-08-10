// Package config handles loading disc configuration from the environment
// and a local .env file.
package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	// TokenEnv is the env var for the bot token.
	TokenEnv = "DISCORD_TOKEN"
	// ServerIDEnv is the env var for the default server.
	ServerIDEnv = "DISCORD_SERVER_ID"
	// EnvFileName is the local env file loaded at startup.
	EnvFileName = ".env"
)

// Config holds the resolved disc configuration.
type Config struct {
	Token    string
	ServerID string
}

// Load reads environment variables (and a local .env file if present) and
// returns the resolved configuration.
func Load() (Config, error) {
	if err := loadEnvFile(); err != nil {
		return Config{}, err
	}
	return Config{
		Token:    os.Getenv(TokenEnv),
		ServerID: os.Getenv(ServerIDEnv),
	}, nil
}

// loadEnvFile parses a local .env file into the process environment without
// overriding already-set variables.
func loadEnvFile() error {
	data, err := os.ReadFile(EnvFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", EnvFileName, err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return nil
}

// SetEnvVar writes or updates a variable in the local .env file.
func SetEnvVar(key, value string) error {
	lines := []string{}
	if data, err := os.ReadFile(EnvFileName); err == nil {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		k, _, ok := strings.Cut(trimmed, "=")
		if ok && strings.TrimSpace(k) == key {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+"="+value)
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(EnvFileName, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", EnvFileName, err)
	}
	_ = os.Setenv(key, value)
	return nil
}
