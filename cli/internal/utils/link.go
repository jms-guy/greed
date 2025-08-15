package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Automatically opens Link URL from client
func OpenLink(system, link string) error {
	if isRunningInContainer() {
		fmt.Printf("\nPlease open this link in your browser:\n%s\n\n", link)
		return nil
	}

	if system == "linux" && IsWSL() {
		// Windows default browser for WSL
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

// Checks for data relating to containers
func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "docker") || strings.Contains(string(data), "containerd")
}

// Detects WSL running
func IsWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}
