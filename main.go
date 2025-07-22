package main

import (
	"fmt"
	"os"
	"steam-account-switcher-win/commands"
)

// main.go - Steam account switcher for Windows
// Provides CLI to list, switch, and show current Steam accounts by editing loginusers.vdf and registry.
//
// Note: Windows-only. Requires registry access and file system permissions.

// main parses CLI arguments and dispatches to the appropriate action (list, current, switch).
func main() {
	action := ""

	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "list":
		// List all accounts in the steam path
		accounts := commands.ListAccounts()

		fmt.Println("Accounts:")
		for _, account := range accounts {
			fmt.Println(" - " + account)
		}
	case "current":
		commands.DisplayCurrentAccount()
	case "switch":
		startSteam := false
		var customCommand *string
		// Check if the account name is provided
		if len(os.Args) < 3 {
			fmt.Println("Please provide an account name to switch to.")
			fmt.Println("Usage: steam-account-switcher switch <account_name> [-y]")
			fmt.Println("Options:")
			fmt.Println("  -y: Start Steam automatically after switching")
			fmt.Println("  -c: Custom command to run after switching")
			return
		}

		// Check if the -y option is provided
		if len(os.Args) > 3 && os.Args[3] == "-y" {
			startSteam = true
		}

		// Check if the -c option is provided
		if len(os.Args) > 4 && os.Args[4] == "-c" {
			if len(os.Args) > 5 {
				cmd := os.Args[5]
				customCommand = &cmd
			}
		}

		commands.SwitchAccount(os.Args[2], startSteam, customCommand)
	case "version":
		commands.DisplayCurrentVersion()
	default:
		fmt.Println("No arguments provided. Please use 'list' or 'switch' or 'current'.")
	}
} 