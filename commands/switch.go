package commands

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"steam-account-switcher-win/utils"
)

// switchAccount switches to the given account, updates files/registry, and optionally starts Steam.
func SwitchAccount(accountName string, startSteam bool, customCommand *string) {
	accounts := ListAccounts()

	// Check if the account name is in the list of accounts
	if !slices.Contains(accounts, accountName) {
		fmt.Println("Account not found. Please try to login with this account to Steam first.")
		return
	}

	fmt.Println("Switching to account:", accountName)

	fmt.Println("Kill any running Steam processes...")
	utils.KillSteamProcesses()

	fmt.Println("Updating loginusers.vdf...")
	utils.SetActiveAccountIntoLoginUsersFile(accountName)

	fmt.Println("Updating AutoLoginUser registry value...")
	utils.UpdateAutoLoginUserRegistry(accountName)

	steamShouldBeStarted := startSteam

	// Prompt user if not auto-starting Steam
	if(!steamShouldBeStarted) {
		fmt.Println("Should I start Steam now? (y/n)")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "y" {
			steamShouldBeStarted = true
		} else {
			fmt.Println("Steam will not be started. You can start it manually.")
		}
	}

	if(steamShouldBeStarted) {
		fmt.Println("Starting Steam...")
		utils.LaunchSteam(customCommand)
		time.Sleep(3000 * time.Millisecond)
		fmt.Println("Waiting for Steam to open...")
	}

	fmt.Println("Account switched successfully !")
}