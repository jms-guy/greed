package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

//Automatically opens Link URL from client
func OpenLink(system, link string) error {
	if system == "linux" && IsWSL() {
		// Use Windows default browser for WSL
		return exec.Command("cmd.exe", "/c", "start", link).Run()
	}

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

//Detects WSL running
func IsWSL() bool {
    data, err := os.ReadFile("/proc/version")
    if err != nil {
        return false
    }
    return strings.Contains(strings.ToLower(string(data)), "microsoft")
}