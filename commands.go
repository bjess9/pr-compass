package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		err := openInBrowser(url)
		if err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func openInBrowser(url string) error {
	var cmd *exec.Cmd

	if isWSL() {
		fmt.Println("WSL")
		cmd = exec.Command("explorer.exe", url)
	} else {
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("cmd", "/C", "start", "", "/B", url)
		case "darwin":
			cmd = exec.Command("open", url)
		default:
			return fmt.Errorf("unsupported platform")
		}
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}
	return nil
}

func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	if _, err := exec.LookPath("wslpath"); err == nil {
		return true
	}

	if content, err := exec.Command("uname", "-r").Output(); err == nil {
		if strings.Contains(strings.ToLower(string(content)), "microsoft") {
			return true
		}
	}

	if os.Getenv("WSL_INTEROP") != "" || os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	return false
}
