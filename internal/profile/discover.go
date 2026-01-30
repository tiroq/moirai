package profile

import (
	"path/filepath"
	"sort"
	"strings"

	"moirai/internal/util"
)

const profilePrefix = "oh-my-opencode.json."

// ProfileInfo describes a discovered profile.
type ProfileInfo struct {
	Name string
	Path string
}

// DiscoverProfiles returns the profiles found in dir.
func DiscoverProfiles(dir string) ([]ProfileInfo, error) {
	entries, err := util.ListDir(dir)
	if err != nil {
		return nil, err
	}

	profiles := make([]ProfileInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, ".bak.") {
			continue
		}
		if !strings.HasPrefix(name, profilePrefix) {
			continue
		}
		profileName := strings.TrimPrefix(name, profilePrefix)
		if profileName == "" {
			continue
		}
		profiles = append(profiles, ProfileInfo{
			Name: profileName,
			Path: filepath.Join(dir, name),
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles, nil
}
