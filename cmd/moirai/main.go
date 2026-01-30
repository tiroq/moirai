package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	default:
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: moirai list")
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
