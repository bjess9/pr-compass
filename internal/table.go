package internal

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/google/go-github/v55/github"
	"github.com/olekukonko/tablewriter"
)

func DisplayPRs(prs []*github.PullRequest) {
	table := tablewriter.NewWriter(os.Stdout)
	yellow := color.New(color.FgYellow).SprintFunc()

	sort.Slice(prs, func(i, j int) bool {
		return prs[i].GetCreatedAt().Time.After(prs[j].GetCreatedAt().Time)
	})

	table.SetHeader([]string{
		"REPO",
		"NUMBER",
		"TITLE",
		"AUTHOR",
		"TIME SINCE OPENED",
	})

	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
	})

	for _, pr := range prs {
		repoName := pr.GetBase().GetRepo().GetName()
		number := fmt.Sprintf("#%d", pr.GetNumber())
		title := yellow(pr.GetTitle())
		author := pr.GetUser().GetLogin()
		createdAt := pr.GetCreatedAt().Time
		timeSinceOpened := humanizeTimeSince(createdAt)

		table.Append([]string{
			repoName,
			number,
			title,
			author,
			timeSinceOpened,
		})
	}

	table.SetBorder(true)
	table.Render()
}
