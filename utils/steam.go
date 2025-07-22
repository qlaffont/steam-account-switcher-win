package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// getSteamPath returns the Steam install path, checking registry first, then STEAM_PATH env var. Exits if not found.
func getSteamPath() string {
	steamPath := GetSteamPathFromRegistry()

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


// getLoginUsersFilePath returns the path to Steam's loginusers.vdf file.
func GetLoginUsersFilePath() string {
	steamPath := getSteamPath()
	return steamPath + "\\config\\loginusers.vdf"
}


// launchSteam starts Steam.exe from the given path.
func LaunchSteam(customCommand *string) {
	if customCommand != nil {
		// Split the command into program and arguments
		parts := strings.Fields(*customCommand)
		// Replace STEAM_PATH placeholder in any part
		for i, part := range parts {
			parts[i] = strings.Replace(part, "STEAM_PATH", getSteamPath()+"\\steam.exe", 1)
		}

		fmt.Println("Executing custom command:", strings.Join(parts, " "))
		
		cmd := exec.Command(parts[0], parts[1:]...)
		// Run the command and capture combined stdout and stderr
		output, err := cmd.CombinedOutput()
		fmt.Println(string(output))
		if err != nil {
			fmt.Println("Error executing custom command:", err)
		}
	} else {
		err := exec.Command(getSteamPath() + "\\steam.exe").Start()
		if err != nil {
			fmt.Println("Error starting Steam:", err)
		}
	}
}

// setActiveAccountIntoLoginUsersFile updates loginusers.vdf to set the given account as active (MostRecent, AllowAutoLogin, etc).
func SetActiveAccountIntoLoginUsersFile(accountName string) {
	loginUsersFile := GetLoginUsersFilePath()

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