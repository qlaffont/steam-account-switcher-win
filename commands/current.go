package commands

import (
	"fmt"
	"steam-account-switcher-win/utils"
)


func DisplayCurrentAccount() {
	fmt.Println(utils.GetAutoLoginUserRegistry())
}