// Package cmd provides the entry point for the repoman CLI.
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/liffiton/repoman/internal/git"
	"github.com/liffiton/repoman/internal/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var noFetch bool

func init() {
	statusCmd.Flags().BoolVarP(&noFetch, "no-fetch", "n", false, "Do not fetch from remote")
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

		manager := git.NewManager(20)
		var gitRepos []git.RepoInfo
		for _, r := range ctx.Repos {
			gitRepos = append(gitRepos, git.RepoInfo{
				Name: r.Name,
				Path: r.Name,
			})
		}

		repoStatuses := manager.StatusAllCtx(cmd.Context(), gitRepos, !noFetch, func() {
			bar.Increment()
		})

		fmt.Println() // New line after progress bar

		results := make([][]string, len(repoStatuses)+1)
		results[0] = []string{"STUDENT/REPO", "BRANCH", "LAST COMMIT", "LOCAL STATUS", "SYNC STATE"}

		for i, s := range repoStatuses {
			if s.Error != nil {
				results[i+1] = []string{s.Name, "ERROR", "-", pterm.Red(s.Error.Error()), "-"}
				continue
			}

			branch := s.Branch
			if branch == "" {
				branch = "-"
			}

			results[i+1] = []string{
				s.Name,
				branch,
				formatCommitTime(s.LastCommit),
				colorStatus(s.Status),
				colorSyncState(s.SyncState),
			}
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(results).Render()

		return nil
	},
}

func colorStatus(status string) string {
	if status == "Clean" {
		return pterm.Green(status)
	}
	if strings.HasPrefix(status, "Error: ") {
		return pterm.Red(status)
	}
	if status == "Missing" {
		return pterm.Red(status)
	}
	if strings.Contains(status, "modified") {
		return pterm.Yellow(status)
	}
	return status
}

func colorSyncState(state string) string {
	if state == "" {
		return "-"
	}
	if state == "Synced" {
		return pterm.Green(state)
	}
	if strings.Contains(state, "Error") {
		return pterm.Red(state)
	}
	if strings.Contains(state, "Stale") || state == "Unknown" || state == "No Upstream" {
		return pterm.Yellow(state)
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
