package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/sys/windows/registry"
)

// main.go - Steam account switcher for Windows
// Provides CLI to list, switch, and show current Steam accounts by editing loginusers.vdf and registry.
//
// Note: Windows-only. Requires registry access and file system permissions.
func getSteamPathFromRegistry() string {
	// Open the Steam registry key
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println("Error opening registry key:", err)
		return ""
	}
	defer k.Close()

	// Get the InstallPath value
	steamPath, _, err := k.GetStringValue("InstallPath")
	if err != nil {
		fmt.Println("Error reading registry value:", err)
		return ""
	}

	return steamPath
}

// getSteamPath returns the Steam install path, checking registry first, then STEAM_PATH env var. Exits if not found.
func getSteamPath() string {
	steamPath := getSteamPathFromRegistry()

	if steamPath == "" {
		// Try to get the steam path from the environment variable if not found in registry; kill the program if not found
		steamPath = os.Getenv("STEAM_PATH")
		if steamPath == "" {
			fmt.Println("Steam path not found in registry or environment variable")
			os.Exit(1)
		}
	}

	return steamPath
}

// getLoginUsersFilePath returns the path to Steam's loginusers.vdf file.
func getLoginUsersFilePath() string {
	steamPath := getSteamPath()
	return steamPath + "\\config\\loginusers.vdf"
}

// listAccounts parses loginusers.vdf and returns all Steam account names found.
func listAccounts() []string {
	loginUsersFile := getLoginUsersFilePath()

	// Read the file
	loginUsers, err := os.ReadFile(loginUsersFile)
	if err != nil {
		fmt.Println("Error reading login users file:", err)
		return []string{}
	}

	// Convert to string for easier parsing
	loginUsersStr := string(loginUsers)

	// Find the "users" section
	usersStart := strings.Index(loginUsersStr, "\"users\"")
	if usersStart == -1 {
		return []string{}
	}

	var accounts []string
	
	// Find all AccountName values using regex
	re := regexp.MustCompile(`"AccountName"\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(loginUsersStr, -1)

	// Extract account names from matches
	for _, match := range matches {
		if len(match) > 1 {
			accounts = append(accounts, match[1])
		}
	}

	return accounts
}

// getCurrentAccount returns the currently set AutoLoginUser from the registry, or empty string on error.
func getCurrentAccount() string {
	// Open the registry key
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println("Error opening registry key:", err)
		return ""
	}

	// Get the AutoLoginUser value
	AutoLoginUser, _, err := k.GetStringValue("AutoLoginUser")
	if err != nil {
		fmt.Println("Error getting registry value:", err)
		return ""
	}

	return AutoLoginUser
}

// killSteamProcesses kills all running steam.exe processes (forcefully).
func killSteamProcesses() {
	// Kill all steam processes
	exec.Command("taskkill", "/F", "/IM", "steam.exe").Run()
}

// parseUserBlock parses a user block (between braces) into a map of key-value pairs.
func parseUserBlock(block string) map[string]string {
	result := make(map[string]string)
	// Match lines like: "Key"    "Value"
	re := regexp.MustCompile(`"([^"]+)"\s+"([^"]+)"`)
	matches := re.FindAllStringSubmatch(block, -1)
	for _, m := range matches {
		if len(m) == 3 {
			result[m[1]] = m[2]
		}
	}
	return result
}

// formatUserBlock pretty-prints a user block with correct tab alignment for loginusers.vdf.
func formatUserBlock(user map[string]string) string {
	// List of keys in the order they appear in test.vdf
	keys := []string{
		"AccountName",
		"PersonaName",
		"RememberPassword",
		"WantsOfflineMode",
		"SkipOfflineModeWarning",
		"AllowAutoLogin",
		"MostRecent",
		"Timestamp",
	}
	var b strings.Builder
	for _, key := range keys {
		if val, ok := user[key]; ok {
			b.WriteString("\t\t\"") // two tabs
			b.WriteString(key)
			b.WriteString("\"\t\t\"")
			b.WriteString(val)
			b.WriteString("\"\n")
		}
	}
	return b.String()
}

// setActiveAccountIntoLoginUsersFile updates loginusers.vdf to set the given account as active (MostRecent, AllowAutoLogin, etc).
func setActiveAccountIntoLoginUsersFile(accountName string) {
	loginUsersFile := getLoginUsersFilePath()

	// Read the file
	loginUsers, err := os.ReadFile(loginUsersFile)
	if err != nil {
		fmt.Println("Error reading login users file:", err)
		return
	}

	loginUsersStr := string(loginUsers)

	// Find the "users" section
	usersStart := strings.Index(loginUsersStr, "\"users\"")
	if usersStart == -1 {
		return
	}

	// Find the opening brace for the users section
	openBrace := strings.Index(loginUsersStr[usersStart:], "{")
	if openBrace == -1 {
		return
	}
	openBrace += usersStart

	// Find the closing brace for the users section by brace counting
	braceCount := 0
	end := openBrace
	for i := openBrace; i < len(loginUsersStr); i++ {
		if loginUsersStr[i] == '{' {
			braceCount++
		} else if loginUsersStr[i] == '}' {
			braceCount--
			if braceCount == 0 {
				end = i
				break
			}
		}
	}

	usersSection := loginUsersStr[openBrace+1 : end] // exclude braces

	var newUsersSection strings.Builder
	pos := 0
	first := true
	for pos < len(usersSection) {
		// Find the next user id
		idStart := strings.Index(usersSection[pos:], "\"")
		if idStart == -1 {
			break
		}
		idStart += pos
		idEnd := strings.Index(usersSection[idStart+1:], "\"")
		if idEnd == -1 {
			break
		}
		idEnd += idStart + 1
		userID := usersSection[idStart+1 : idEnd]

		// Find the opening brace for this user
		userOpen := strings.Index(usersSection[idEnd:], "{")
		if userOpen == -1 {
			break
		}
		userOpen += idEnd

		// Find the matching closing brace for this user
		braceCount := 0
		userClose := userOpen
		for i := userOpen; i < len(usersSection); i++ {
			if usersSection[i] == '{' {
				braceCount++
			} else if usersSection[i] == '}' {
				braceCount--
				if braceCount == 0 {
					userClose = i
					break
				}
			}
		}

		userContent := usersSection[userOpen+1 : userClose]
		userMap := parseUserBlock(userContent)

		// Set active flags for the selected account, clear for others
		if userMap["AccountName"] == accountName {
			userMap["RememberPassword"] = "1"
			userMap["MostRecent"] = "1"
			userMap["AllowAutoLogin"] = "1"
			userMap["Timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
		} else {
			userMap["MostRecent"] = "0"
			userMap["AllowAutoLogin"] = "0"
		}

		if !first {
			newUsersSection.WriteString("\n") // single blank line between user blocks
		}
		first = false
		// Write user block with correct indentation
		newUsersSection.WriteString("\t\"")
		newUsersSection.WriteString(userID)
		newUsersSection.WriteString("\"\n\t{\n")
		newUsersSection.WriteString(formatUserBlock(userMap))
		newUsersSection.WriteString("\t}\n")

		// Move to next user (skip whitespace)
		nextPos := userClose + 1
		for nextPos < len(usersSection) && (usersSection[nextPos] == '\n' || usersSection[nextPos] == '\r' || usersSection[nextPos] == ' ' || usersSection[nextPos] == '\t') {
			nextPos++
		}
		pos = nextPos
	}

	// Rebuild the file
	newLoginUsersStr := loginUsersStr[:openBrace+1] + "\n" + newUsersSection.String() + loginUsersStr[end:]

	// Write the file
	err = os.WriteFile(loginUsersFile, []byte(newLoginUsersStr), 0644)
	if err != nil {
		fmt.Println("Error writing login users file:", err)
	}
}

// updateAutoLoginUserRegistry sets the AutoLoginUser registry value to the given account name.
func updateAutoLoginUserRegistry(accountName string) {
	// Open the registry key with write access
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.SET_VALUE)
	if err != nil {
		fmt.Println("Error opening registry key:", err)
		return
	}

	// Set the AutoLoginUser value to the account name
	err = k.SetStringValue("AutoLoginUser", accountName)
	if err != nil {
		fmt.Println("Error setting registry value:", err)
		return
	}

	// Close the registry key
	k.Close()
}

// switchAccount switches to the given account, updates files/registry, and optionally starts Steam.
func switchAccount(accountName string, startSteam bool) {
	accounts := listAccounts()

	// Check if the account name is in the list of accounts
	if !slices.Contains(accounts, accountName) {
		fmt.Println("Account not found. Please try to login with this account to Steam first.")
		return
	}

	fmt.Println("Switching to account:", accountName)

	fmt.Println("Kill any running Steam processes...")
	killSteamProcesses()

	fmt.Println("Waiting for Steam to close...")
	time.Sleep(1000 * time.Millisecond)

	fmt.Println("Updating loginusers.vdf...")
	setActiveAccountIntoLoginUsersFile(accountName)

	fmt.Println("Updating AutoLoginUser registry value...")
	updateAutoLoginUserRegistry(accountName)

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
		steamPath := getSteamPath()
		exec.Command(steamPath + "\\steam.exe").Start()
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("Waiting for Steam to open...")
	}

	fmt.Println("Account switched successfully !")
}

// main parses CLI arguments and dispatches to the appropriate action (list, current, switch).
func main() {
	action := ""

	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "list":
		// List all accounts in the steam path
		accounts := listAccounts()

		fmt.Println("Accounts:")
		for _, account := range accounts {
			fmt.Println(" - " + account)
		}
	case "current":
		fmt.Println(getCurrentAccount())
	case "switch":
		startSteam := false
		// Check if the account name is provided
		if len(os.Args) < 3 {
			fmt.Println("Please provide an account name to switch to.")
			fmt.Println("Usage: steam-account-switcher switch <account_name> [-y]")
			fmt.Println("Options:")
			fmt.Println("  -y: Start Steam automatically after switching")
			return
		}

		// Check if the -y option is provided
		if len(os.Args) > 3 && os.Args[3] == "-y" {
			startSteam = true
		}

		switchAccount(os.Args[2], startSteam)
	default:
		fmt.Println("No arguments provided. Please use 'list' or 'switch' or 'current'.")
	}
} 