package utils

import (
	"fmt"
	"os/exec"
	"time"

	"golang.org/x/sys/windows/registry"
)

func GetSteamPathFromRegistry() string {
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


// getCurrentAccount returns the currently set AutoLoginUser from the registry, or empty string on error.
func GetAutoLoginUserRegistry() string {
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

// updateAutoLoginUserRegistry sets the AutoLoginUser registry value to the given account name.
func UpdateAutoLoginUserRegistry(accountName string) {
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


// killSteamProcesses kills all running steam.exe processes (forcefully).
func KillSteamProcesses() {
	steamIsRunning := true

	for(steamIsRunning) {
		// Kill all steam processes
		err := exec.Command("taskkill", "/F", "/IM", "steam.exe").Run()
		if err != nil {
			// Exit status 128 means no matching processes found
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
				steamIsRunning = false
				continue
			}
			fmt.Println("Error killing Steam process:", err)
		}
		time.Sleep(1000 * time.Millisecond)
		
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq steam.exe")
		err = cmd.Run()
		steamIsRunning = err == nil
	}
}