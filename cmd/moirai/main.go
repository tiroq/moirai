package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"moirai/internal/backup"
	"moirai/internal/link"
	"moirai/internal/profile"
	"moirai/internal/util"
)

const defaultConfigDir = "~/.config/opencode"

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		printHelp()
		return
	}

	switch args[0] {
	case "list":
		if err := runList(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "apply":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai apply <profile>")
			os.Exit(1)
		}
		if err := runApply(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "doctor":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai doctor <profile>")
			os.Exit(1)
		}
		exitCode, err := runDoctor(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	case "backup":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai backup <profile>")
			os.Exit(1)
		}
		if err := runBackup(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "backups":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai backups <profile>")
			os.Exit(1)
		}
		if err := runBackups(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "restore":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai restore <profile> --from <backupPathOrFilename>")
			os.Exit(1)
		}
		restoreFlags := flag.NewFlagSet("restore", flag.ContinueOnError)
		from := restoreFlags.String("from", "", "backup path or filename")
		if err := restoreFlags.Parse(args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := runRestore(args[1], *from); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: moirai list")
	fmt.Println("       moirai apply <profile>")
	fmt.Println("       moirai doctor <profile>")
	fmt.Println("       moirai backup <profile>")
	fmt.Println("       moirai backups <profile>")
	fmt.Println("       moirai restore <profile> --from <backupPathOrFilename>")
	fmt.Println("       moirai help")
}

func runList() error {
	configDir, err := util.ExpandUser(defaultConfigDir)
	if err != nil {
		return err
	}
	configDir = filepath.Clean(configDir)

	profiles, err := profile.DiscoverProfiles(configDir)
	if err != nil {
		return err
	}

	activeName, ok, err := link.ActiveProfile(configDir)
	if err != nil {
		return err
	}

	fmt.Printf("ConfigDir: %s\n", configDir)
	if ok {
		fmt.Printf("Active: %s\n", activeName)
	} else {
		fmt.Println("Active: (none)")
	}
	fmt.Println("Profiles:")
	for _, info := range profiles {
		suffix := ""
		if ok && info.Name == activeName {
			suffix = " *"
		}
		fmt.Printf(" - %s%s\n", info.Name, suffix)
	}
	return nil
}

func runApply(profileName string) error {
	configDir, err := util.ExpandUser(defaultConfigDir)
	if err != nil {
		return err
	}
	configDir = filepath.Clean(configDir)

	if err := link.ApplyProfile(configDir, profileName); err != nil {
		return err
	}
	fmt.Printf("Applied: %s\n", profileName)
	return nil
}

func runDoctor(profileName string) (int, error) {
	configDir, err := util.ExpandUser(defaultConfigDir)
	if err != nil {
		return 1, err
	}
	configDir = filepath.Clean(configDir)

	profilePath := filepath.Join(configDir, fmt.Sprintf("oh-my-opencode.json.%s", profileName))
	cfg, err := profile.LoadProfile(profilePath)
	if err != nil {
		return 1, err
	}

	missing := profile.MissingAgents(cfg, profile.KnownAgents())

	fmt.Printf("Profile: %s\n", profileName)
	fmt.Println("Missing:")
	if len(missing) == 0 {
		fmt.Println(" (none)")
		return 0, nil
	}
	for _, agent := range missing {
		fmt.Printf(" - %s\n", agent)
	}
	return 2, nil
}

func runBackup(profileName string) error {
	configDir, err := util.ExpandUser(defaultConfigDir)
	if err != nil {
		return err
	}
	configDir = filepath.Clean(configDir)

	backupPath, err := backup.BackupProfile(configDir, profileName)
	if err != nil {
		return err
	}
	fmt.Printf("Backup: %s\n", backupPath)
	return nil
}

func runBackups(profileName string) error {
	configDir, err := util.ExpandUser(defaultConfigDir)
	if err != nil {
		return err
	}
	configDir = filepath.Clean(configDir)

	backups, err := backup.ListProfileBackups(configDir, profileName)
	if err != nil {
		return err
	}

	fmt.Println("Backups:")
	if len(backups) == 0 {
		fmt.Println(" (none)")
		return nil
	}
	for _, name := range backups {
		fmt.Printf(" - %s\n", name)
	}
	return nil
}

func runRestore(profileName, from string) error {
	configDir, err := util.ExpandUser(defaultConfigDir)
	if err != nil {
		return err
	}
	configDir = filepath.Clean(configDir)

	preBackupPath, err := backup.RestoreProfileFromBackup(configDir, profileName, from)
	if err != nil {
		return err
	}
	fmt.Printf("Restored: %s\n", profileName)
	fmt.Printf("PreBackup: %s\n", preBackupPath)
	return nil
}
