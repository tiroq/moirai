package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"moirai/internal/app"
	"moirai/internal/backup"
	"moirai/internal/link"
	"moirai/internal/profile"
	"moirai/internal/util"
)

const defaultConfigDir = "~/.config/opencode"

func main() {
	args, globalFlags, err := parseGlobalFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		printHelp()
		return
	}

	configDir, err := util.ExpandUser(defaultConfigDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	configDir = filepath.Clean(configDir)
	var enableAutofillOverride *bool
	if globalFlags.EnableAutofillSet {
		enableAutofillOverride = &globalFlags.EnableAutofill
	}
	appConfig, err := app.LoadConfig(configDir, enableAutofillOverride)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		if err := runList(appConfig); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "apply":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai apply <profile>")
			os.Exit(1)
		}
		if err := runApply(appConfig, args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "doctor":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai doctor <profile>")
			os.Exit(1)
		}
		exitCode, err := runDoctor(appConfig, args[1])
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
		if err := runBackup(appConfig, args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "backups":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai backups <profile>")
			os.Exit(1)
		}
		if err := runBackups(appConfig, args[1]); err != nil {
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
		if err := runRestore(appConfig, args[1], *from); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "diff":
		exitCode, err := runDiff(appConfig, args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	case "autofill":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: moirai autofill <profile> --preset <preset>")
			os.Exit(1)
		}
		autofillFlags := flag.NewFlagSet("autofill", flag.ContinueOnError)
		preset := autofillFlags.String("preset", "", "autofill preset")
		if err := autofillFlags.Parse(args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *preset == "" {
			fmt.Fprintln(os.Stderr, "Usage: moirai autofill <profile> --preset <preset>")
			os.Exit(1)
		}
		exitCode, err := runAutofill(appConfig, args[1], *preset)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	default:
		printHelp()
		os.Exit(1)
	}
}

type globalFlags struct {
	EnableAutofill    bool
	EnableAutofillSet bool
}

func parseGlobalFlags(args []string) ([]string, globalFlags, error) {
	var flags globalFlags
	remaining := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--enable-autofill" {
			flags.EnableAutofill = true
			flags.EnableAutofillSet = true
			continue
		}
		if strings.HasPrefix(arg, "--enable-autofill=") {
			value := strings.TrimPrefix(arg, "--enable-autofill=")
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return nil, flags, fmt.Errorf("invalid value for --enable-autofill: %q", value)
			}
			flags.EnableAutofill = parsed
			flags.EnableAutofillSet = true
			continue
		}
		remaining = append(remaining, arg)
	}
	return remaining, flags, nil
}

func printHelp() {
	fmt.Println("Usage: moirai list")
	fmt.Println("       moirai apply <profile>")
	fmt.Println("       moirai doctor <profile>")
	fmt.Println("       moirai backup <profile>")
	fmt.Println("       moirai backups <profile>")
	fmt.Println("       moirai restore <profile> --from <backupPathOrFilename>")
	fmt.Println("       moirai diff <profile> --against last-backup")
	fmt.Println("       moirai diff --between <profileA> <profileB>")
	fmt.Println("       moirai autofill <profile> --preset <preset>")
	fmt.Println("Global options:")
	fmt.Println("       --enable-autofill")
	fmt.Println("       moirai help")
}

func runList(config app.AppConfig) error {
	profiles, err := profile.DiscoverProfiles(config.ConfigDir)
	if err != nil {
		return err
	}

	activeName, ok, err := link.ActiveProfile(config.ConfigDir)
	if err != nil {
		return err
	}

	fmt.Printf("ConfigDir: %s\n", config.ConfigDir)
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

func runApply(config app.AppConfig, profileName string) error {
	if err := link.ApplyProfile(config.ConfigDir, profileName); err != nil {
		return err
	}
	fmt.Printf("Applied: %s\n", profileName)
	return nil
}

func runDoctor(config app.AppConfig, profileName string) (int, error) {
	profilePath := filepath.Join(config.ConfigDir, fmt.Sprintf("oh-my-opencode.json.%s", profileName))
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

func runBackup(config app.AppConfig, profileName string) error {
	backupPath, err := backup.BackupProfile(config.ConfigDir, profileName)
	if err != nil {
		return err
	}
	fmt.Printf("Backup: %s\n", backupPath)
	return nil
}

func runBackups(config app.AppConfig, profileName string) error {
	backups, err := backup.ListProfileBackups(config.ConfigDir, profileName)
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

func runRestore(config app.AppConfig, profileName, from string) error {
	preBackupPath, err := backup.RestoreProfileFromBackup(config.ConfigDir, profileName, from)
	if err != nil {
		return err
	}
	fmt.Printf("Restored: %s\n", profileName)
	fmt.Printf("PreBackup: %s\n", preBackupPath)
	return nil
}

func runDiff(config app.AppConfig, args []string) (int, error) {
	if len(args) == 0 {
		printDiffHelp()
		return 1, fmt.Errorf("missing diff arguments")
	}
	if args[0] == "--between" {
		if len(args) != 3 {
			printDiffHelp()
			return 1, fmt.Errorf("Usage: moirai diff --between <profileA> <profileB>")
		}
		return runDiffBetween(config, args[1], args[2])
	}

	if len(args) < 2 {
		printDiffHelp()
		return 1, fmt.Errorf("Usage: moirai diff <profile> --against last-backup")
	}
	if args[1] != "--against" || len(args) != 3 || args[2] != "last-backup" {
		printDiffHelp()
		return 1, fmt.Errorf("Usage: moirai diff <profile> --against last-backup")
	}
	return runDiffAgainstLastBackup(config, args[0])
}

func printDiffHelp() {
	fmt.Println("Usage: moirai diff <profile> --against last-backup")
	fmt.Println("       moirai diff --between <profileA> <profileB>")
}

func runDiffAgainstLastBackup(config app.AppConfig, profileName string) (int, error) {
	profilePath := filepath.Join(config.ConfigDir, fmt.Sprintf("oh-my-opencode.json.%s", profileName))
	if _, err := os.Stat(profilePath); err != nil {
		return 1, err
	}

	backupName, ok, err := backup.LatestProfileBackup(config.ConfigDir, profileName)
	if err != nil {
		return 1, err
	}
	if !ok {
		fmt.Printf("No backups found for profile: %s\n", profileName)
		return 2, nil
	}
	backupPath := filepath.Join(config.ConfigDir, backupName)

	diff, err := util.GitDiffNoIndex(backupPath, profilePath)
	if err != nil {
		if errors.Is(err, util.ErrGitNotAvailable) {
			return 1, fmt.Errorf("git is required for diff: %w", err)
		}
		return 1, err
	}
	fmt.Print(diff)
	return 0, nil
}

func runDiffBetween(config app.AppConfig, profileA, profileB string) (int, error) {
	pathA := filepath.Join(config.ConfigDir, fmt.Sprintf("oh-my-opencode.json.%s", profileA))
	pathB := filepath.Join(config.ConfigDir, fmt.Sprintf("oh-my-opencode.json.%s", profileB))

	if _, err := os.Stat(pathA); err != nil {
		return 1, err
	}
	if _, err := os.Stat(pathB); err != nil {
		return 1, err
	}

	diff, err := util.GitDiffNoIndex(pathA, pathB)
	if err != nil {
		if errors.Is(err, util.ErrGitNotAvailable) {
			return 1, fmt.Errorf("git is required for diff: %w", err)
		}
		return 1, err
	}
	fmt.Print(diff)
	return 0, nil
}

func runAutofill(config app.AppConfig, profileName, presetName string) (int, error) {
	if !config.EnableAutofill {
		fmt.Fprintln(os.Stderr, "Autofill is disabled. Enable with --enable-autofill or moirai.json.")
		return 3, nil
	}
	if profileName == "" {
		return 1, fmt.Errorf("profile name is required")
	}

	preset, ok := profile.PresetByName(presetName)
	if !ok {
		return 1, fmt.Errorf("unknown preset: %s", presetName)
	}

	profilePath := filepath.Join(config.ConfigDir, fmt.Sprintf("oh-my-opencode.json.%s", profileName))
	cfg, err := profile.LoadProfile(profilePath)
	if err != nil {
		return 1, err
	}

	changed := profile.ApplyAutofill(cfg, profile.KnownAgents(), preset)
	if !changed {
		fmt.Printf("No changes needed for profile: %s\n", profileName)
		return 0, nil
	}

	backupPath, err := backup.BackupProfile(config.ConfigDir, profileName)
	if err != nil {
		return 1, err
	}
	if err := profile.SaveProfileAtomic(profilePath, cfg); err != nil {
		return 1, err
	}

	fmt.Printf("Autofilled: %s\n", profileName)
	fmt.Printf("Backup: %s\n", backupPath)
	return 0, nil
}
