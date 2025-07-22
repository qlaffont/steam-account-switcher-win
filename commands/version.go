package commands

import "fmt"

const version = "1.1.0"

func DisplayCurrentVersion() {
	fmt.Println("Current version: " + version)
}