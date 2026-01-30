package link

import (
	"os"
	"path/filepath"
	"strings"
)

const activeFileName = "oh-my-opencode.json"
const profilePrefix = "oh-my-opencode.json."

// ActiveProfile reports the active profile name if the active file is a symlink.
func ActiveProfile(dir string) (string, bool, error) {
	activePath := filepath.Join(dir, activeFileName)
	info, err := os.Lstat(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return "", false, nil
	}

	linkTarget, err := os.Readlink(activePath)
	if err != nil {
		return "", false, err
	}

	base := filepath.Base(linkTarget)
	if strings.Contains(base, ".bak.") {
		return "", false, nil
	}
	if !strings.HasPrefix(base, profilePrefix) {
		return "", false, nil
	}
	name := strings.TrimPrefix(base, profilePrefix)
	if name == "" {
		return "", false, nil
	}

	fullTarget := linkTarget
	if !filepath.IsAbs(fullTarget) {
		fullTarget = filepath.Join(dir, linkTarget)
	}
	stat, err := os.Stat(fullTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	if stat.IsDir() {
		return "", false, nil
	}

	return name, true, nil
}
