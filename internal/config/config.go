// Package config handles loading disc configuration from the environment,
// a local .env file, and the global disc config directory.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// TokenEnv is the env var for the bot token.
	TokenEnv = "DISCORD_TOKEN"
	// ServerIDEnv is the env var for the default server.
	ServerIDEnv = "DISCORD_SERVER_ID"
	// EnvFileName is the local env file loaded at startup.
	EnvFileName = ".env"
	// ConfigLocationEnv is the env var that overrides the stored preference.
	ConfigLocationEnv = "DISC_CONFIG_LOCATION"
)

// Config holds the resolved disc configuration.
type Config struct {
	Token    string
	ServerID string
}

// Prefs holds the stored user preferences.
type Prefs struct {
	ServerConfigLocation string `json:"server_config_location"` // "local" or "global"
}

const prefsFile = "prefs.json"

// GlobalDir returns the disc config directory (~/.config/disc).
func GlobalDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config dir: %w", err)
	}
	return filepath.Join(dir, "disc"), nil
}

// ServerDir returns the directory for a given server.
func ServerDir(serverID string) (string, error) {
	gd, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(gd, "servers", serverID), nil
}

// ServerEnvFile returns the env file path for a given server.
func ServerEnvFile(serverID string) (string, error) {
	sd, err := ServerDir(serverID)
	if err != nil {
		return "", err
	}
	return filepath.Join(sd, "env"), nil
}

// ServerConfigFile returns the server.json path for a given server.
func ServerConfigFile(serverID string) (string, error) {
	sd, err := ServerDir(serverID)
	if err != nil {
		return "", err
	}
	return filepath.Join(sd, "server.json"), nil
}

// PrefsPath returns the prefs.json path.
func PrefsPath() (string, error) {
	gd, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(gd, prefsFile), nil
}

// LoadPrefs reads the stored preferences.
func LoadPrefs() (*Prefs, error) {
	path, err := PrefsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Prefs{}, nil
		}
		return nil, fmt.Errorf("failed to read prefs: %w", err)
	}
	var p Prefs
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse prefs: %w", err)
	}
	return &p, nil
}

// SavePrefs writes the preferences.
func SavePrefs(p *Prefs) error {
	path, err := PrefsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode prefs: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write prefs: %w", err)
	}
	return nil
}

// loadEnvFile parses a local .env file into the process environment without
// overriding already-set variables.
func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", path, err)
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

// SingleServerDir returns the server ID if exactly one servers/<sid> directory
// exists in the global config dir. Returns "" if zero or more than one.
func SingleServerDir() string {
	gd, err := GlobalDir()
	if err != nil {
		return ""
	}
	serversPath := filepath.Join(gd, "servers")
	entries, err := os.ReadDir(serversPath)
	if err != nil {
		return ""
	}
	var found string
	for _, e := range entries {
		if e.IsDir() {
			if found != "" {
				return "" // more than one
			}
			found = e.Name()
		}
	}
	return found
}

// Load reads environment variables (local .env, then global server env) and
// returns the resolved configuration. Global env is only loaded when the
// server is unambiguous: set via DISCORD_SERVER_ID, or only one servers/<sid>
// directory exists.
func Load() (Config, error) {
	if err := loadEnvFile(EnvFileName); err != nil {
		return Config{}, err
	}

	sid := os.Getenv(ServerIDEnv)
	if sid == "" {
		sid = SingleServerDir()
	}

	if sid != "" {
		_ = loadEnvFileForServer(sid)
	}

	return Config{
		Token:    os.Getenv(TokenEnv),
		ServerID: os.Getenv(ServerIDEnv),
	}, nil
}

// LoadForServer loads the global env for a specific server (e.g. when
// --server is passed), without overriding already-set variables.
func LoadForServer(serverID string) error {
	return loadEnvFileForServer(serverID)
}

// loadEnvFileForServer loads the env file for a given server ID.
func loadEnvFileForServer(serverID string) error {
	path, err := ServerEnvFile(serverID)
	if err != nil {
		return err
	}
	return loadEnvFile(path)
}

// SetEnvVar writes or updates a variable in the local .env file.
func SetEnvVar(key, value string) error {
	return setEnvVarIn(EnvFileName, key, value)
}

// SetServerEnvVar writes or updates a variable in a server's env file.
func SetServerEnvVar(serverID, key, value string) error {
	path, err := ServerEnvFile(serverID)
	if err != nil {
		return err
	}
	return setEnvVarIn(path, key, value)
}

func setEnvVarIn(envPath, key, value string) error {
	lines := []string{}
	if data, err := os.ReadFile(envPath); err == nil {
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
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return fmt.Errorf("failed to create dir for %s: %w", envPath, err)
	}
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", envPath, err)
	}
	_ = os.Setenv(key, value)
	return nil
}

// ResolveServerDir returns the server dir for a given server ID, creating it
// if create is true.
func ResolveServerDir(serverID string, create bool) (string, error) {
	sd, err := ServerDir(serverID)
	if err != nil {
		return "", err
	}
	if create {
		if err := os.MkdirAll(sd, 0o755); err != nil {
			return "", fmt.Errorf("failed to create server dir: %w", err)
		}
	}
	return sd, nil
}
