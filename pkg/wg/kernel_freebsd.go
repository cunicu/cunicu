package wg

import (
	"os/exec"
)

func KernelModuleExists() bool {
	devName := "must-not-exist"

	if err := exec.Command("ifconfig", "wg", "create", "name", devName).Run(); err != nil {
		return false
	}

	if err := exec.Command("ifconfig", devName, "destroy").Run(); err != nil {
		return false
	}

	return true
}
