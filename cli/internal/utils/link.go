package utils

import (
	"fmt"
	"os/exec"
)

//Automatically opens Link URL from client
func OpenLink(system, link string) error {

	switch system {
	case "linux":
		return exec.Command("xdg-open", link).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", link).Run()
	case "darwin":
		return exec.Command("open", link).Run()
	default:
		return fmt.Errorf("operating system %q not supported for automatic link flow", system)
	}
}
