package cmd

import (
	"fmt"
	"os"
	"strings"
)

func findChannelInConfig(cfg *serverConfigFile, id string) int {
	for i := range cfg.Channels {
		if cfg.Channels[i].ID == id {
			return i
		}
	}
	return -1
}

func applyConfigOverwrites(ce *channelExport, allows, denies []string) error {
	targets, err := collectPermTargets(allows, denies)
	if err != nil {
		return err
	}
	idx := map[string]int{}
	for i, o := range ce.Overwrites {
		idx[o.Role] = i
	}
	for role, t := range targets {
		i, exists := idx[role]
		ow := permissionOverwriteExport{Role: role}
		if exists {
			ow = ce.Overwrites[i]
		}
		if t.allowChanged {
			ow.Allow = t.allow
		}
		if t.denyChanged {
			ow.Deny = t.deny
		}
		if exists {
			ce.Overwrites[i] = ow
		} else {
			ce.Overwrites = append(ce.Overwrites, ow)
		}
	}
	return nil
}

func removeConfigOverwrite(ce *channelExport, roles []string) {
	removed := map[string]bool{}
	for _, r := range roles {
		removed[r] = true
	}
	keep := ce.Overwrites[:0]
	for _, o := range ce.Overwrites {
		if !removed[o.Role] {
			keep = append(keep, o)
		}
	}
	ce.Overwrites = keep
}

func splitPermissions(s string) ([]string, error) {
	permBits := permissionBitsByName()
	var perms []string
	for _, part := range strings.Split(s, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, ok := permBits[name]; !ok {
			return nil, fmt.Errorf("unknown permission '%s'; see 'disc role perm list'", name)
		}
		perms = append(perms, name)
	}
	return perms, nil
}

func findRoleInConfig(cfg *serverConfigFile, id string) int {
	for i := range cfg.Roles {
		if cfg.Roles[i].ID == id {
			return i
		}
	}
	return -1
}

func loadConfigFileForEdit(file, serverHint string, create bool) (*serverConfigFile, string, error) {
	path, err := resolveConfigFile(file, serverHint)
	if err != nil {
		return nil, "", err
	}
	var cfg serverConfigFile
	if err := readJSONFile(path, &cfg); err != nil {
		if create && os.IsNotExist(err) {
			if err := writeConfigFile(&cfg, path); err != nil {
				return nil, "", err
			}
			return &cfg, path, nil
		}
		return nil, "", err
	}
	return &cfg, path, nil
}
