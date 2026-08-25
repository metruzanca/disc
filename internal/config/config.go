package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func GlobalDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config dir: %w", err)
	}
	return filepath.Join(dir, "disc"), nil
}

func ServerDir(serverID string) (string, error) {
	gd, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(gd, "servers", serverID), nil
}

func ServerConfigFile(serverID string) (string, error) {
	sd, err := ServerDir(serverID)
	if err != nil {
		return "", err
	}
	return filepath.Join(sd, "server.json"), nil
}

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
