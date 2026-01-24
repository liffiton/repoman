// Package cmd provides the entry point for the repoman CLI.
package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/liffiton/repoman/internal/api"
	"github.com/liffiton/repoman/internal/git"
	"github.com/liffiton/repoman/internal/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all student repositories in the workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := loadWorkspaceContext()
		if err != nil {
			return err
		}

		ui.PrintHeader("Status for " + pterm.Bold.Sprintf("%s - %s", ctx.Wcfg.CourseName, ctx.Wcfg.AssignmentName))
		if ctx.OrigDir != ctx.Wcfg.Root {
			ui.Dim.Printf("Workspace: %s\n", ctx.Wcfg.Root)
		}
		pterm.Println()

		bar, _ := ui.Progressbar.WithTotal(len(ctx.Repos)).WithTitle("Checking status").Start()

		var wg sync.WaitGroup
		results := make([][]string, len(ctx.Repos)+1)
		results[0] = []string{"STUDENT/REPO", "BRANCH", "LAST COMMIT", "LOCAL STATUS", "SYNC STATE"}

		sem := make(chan struct{}, 10) // Concurrency limit for status checks

		for i, repo := range ctx.Repos {
			wg.Add(1)
			go func(i int, r api.Repo) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				localPath := r.Name
				if _, err := os.Stat(localPath); os.IsNotExist(err) {
					results[i+1] = []string{r.Name, "-", "-", pterm.Red("Missing"), "-"}
					bar.Increment()
					return
				}

				// Optional: Fetch to get accurate sync state.
				// Since it's 10-30 repos, we'll do it.
				_ = git.Fetch(localPath)

				branch, status, err := git.GetStatus(localPath)
				if err != nil {
					results[i+1] = []string{r.Name, "ERROR", "-", pterm.Red(err.Error()), "-"}
					bar.Increment()
					return
				}

				syncState, _ := git.GetSyncState(localPath)
				lastCommit, _ := git.GetLastCommitTime(localPath)
				results[i+1] = []string{
					r.Name,
					branch,
					formatCommitTime(lastCommit),
					colorStatus(status),
					colorSyncState(syncState),
				}
				bar.Increment()
			}(i, repo)
		}

		wg.Wait()
		fmt.Println() // New line after progress bar

		_ = pterm.DefaultTable.WithHasHeader().WithData(results).Render()

		return nil
	},
}

func colorStatus(status string) string {
	if status == "Clean" {
		return pterm.Green(status)
	}
	if strings.Contains(status, "modified") {
		return pterm.Yellow(status)
	}
	return status
}

func colorSyncState(state string) string {
	if state == "Synced" {
		return pterm.Green(state)
	}
	if strings.HasPrefix(state, "Behind") {
		return pterm.Red(state)
	}
	if strings.HasPrefix(state, "Ahead") {
		return pterm.Blue(state)
	}
	if strings.HasPrefix(state, "Diverged") {
		return pterm.Magenta(state)
	}
	return state
}

func formatCommitTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	local := t.Local()
	now := time.Now().Local()

	dateStr := ""
	if local.Year() == now.Year() && local.YearDay() == now.YearDay() {
		dateStr = "today     "
	} else if yesterday := now.AddDate(0, 0, -1); local.Year() == yesterday.Year() && local.YearDay() == yesterday.YearDay() {
		dateStr = "yesterday "
	} else {
		dateStr = local.Format("2006-01-02")
	}

	return fmt.Sprintf("%s %s", dateStr, local.Format("15:04"))
}
