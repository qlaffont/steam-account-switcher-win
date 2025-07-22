package commands

import (
	"fmt"
	"os"
	"regexp"
	"steam-account-switcher-win/utils"
	"strings"
)

// listAccounts parses loginusers.vdf and returns all Steam account names found.
func ListAccounts() []string {
	loginUsersFile := utils.GetLoginUsersFilePath()

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