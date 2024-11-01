package internal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	gh "github.com/google/go-github/v55/github"

	"github.com/charmbracelet/bubbles/table"
	"golang.org/x/term"
)

func IsWSL() bool {
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

func createTableColumns() []table.Column {
	totalWidth := getTerminalWidth() - 4
	return []table.Column{
		{Title: "PR Name", Width: totalWidth / 3},
		{Title: "Author", Width: totalWidth / 6},
		{Title: "Repository", Width: totalWidth / 3},
		{Title: "Time Open", Width: totalWidth / 6},
	}
}

func createTableRows(prs []*gh.PullRequest) []table.Row {
	rows := make([]table.Row, len(prs))
	for i, pr := range prs {
		prNumber := fmt.Sprintf("#%d", pr.GetNumber())
		prName := prNumber + " " + pr.GetTitle()
		author := pr.GetUser().GetLogin()
		repo := pr.GetBase().GetRepo().GetFullName()
		timeSinceOpened := humanizeTimeSince(pr.GetCreatedAt().Time)

		row := table.Row{
			prName,
			author,
			repo,
			timeSinceOpened,
		}

		rows[i] = row
	}
	return rows
}

func humanizeTimeSince(t time.Time) string {
	duration := time.Since(t)
	if duration.Hours() < 24 {
		if duration.Hours() >= 1 {
			return fmt.Sprintf("%.0fh", duration.Hours())
		}
		return fmt.Sprintf("%.0fm", duration.Minutes())
	}
	return fmt.Sprintf("%.0fd", duration.Hours()/24)
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}
	return width
}

func loadingView() string {
	return "Loading PRs...\nPress 'q' to quit."
}

func errorView(err error) string {
	return fmt.Sprintf("Error: %v\nPress 'q' to quit.", err)
}
