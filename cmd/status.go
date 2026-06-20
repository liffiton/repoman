// Package cmd provides the entry point for the repoman CLI.
package cmd

import (
	"fmt"
	"math"
	"sort"
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

		sort.Slice(repoStatuses, func(i, j int) bool {
			iBad := repoStatuses[i].Status == git.StatusMissing || repoStatuses[i].Status == git.StatusError || repoStatuses[i].Error != nil
			jBad := repoStatuses[j].Status == git.StatusMissing || repoStatuses[j].Status == git.StatusError || repoStatuses[j].Error != nil
			if iBad != jBad {
				return !iBad
			}
			return repoStatuses[i].Name < repoStatuses[j].Name
		})

		fmt.Println() // New line after progress bar

		maxCommits := 0
		for _, s := range repoStatuses {
			maxCommits = max(maxCommits, s.CommitCount)
		}

		results := make([][]string, len(repoStatuses)+1)
		results[0] = []string{"STUDENT/REPO", "BRANCH", "COMMITS", "LAST COMMIT", "LOCAL STATUS", "SYNC STATE"}

		for i, s := range repoStatuses {
			if s.Error != nil {
				results[i+1] = []string{
					s.Name,
					"ERROR",
					dimPlaceholder(7),
					dimPlaceholder(),
					pterm.Red(s.Error.Error()),
					dimPlaceholder(),
				}
				continue
			}

			commits := formatCommitCount(s.CommitCount, maxCommits)
			if s.Status == git.StatusMissing {
				commits = dimPlaceholder(7)
			}

			branch := s.Branch
			if branch == "" {
				branch = dimPlaceholder()
			}

			results[i+1] = []string{
				s.Name,
				branch,
				commits,
				formatCommitTime(s.LastCommit),
				colorStatus(s.Status),
				colorSyncState(s.SyncState),
			}
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(results).Render()

		return nil
	},
}

func dimPlaceholder(width ...int) string {
	dash := "-"
	if len(width) > 0 {
		dash = fmt.Sprintf("%*s", width[0], "-")
	}
	return pterm.NewRGB(105, 105, 105).Sprintf("%s", dash)
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
	if state == "" || state == "-" {
		return dimPlaceholder()
	}
	return state
}

func formatCommitCount(count, maxCommits int) string {
	// right align ...
	formatted := fmt.Sprintf("%7d", count)
	// ... and color based on # commits compared to max (few commits: darkish orange)
	if count == 0 {
		return pterm.NewRGB(230, 80, 80).Sprintf("%s", formatted)
	}
	t := math.Log(float64(count)+0.1) / math.Log(float64(maxCommits)+0.1)
	r := uint8(math.Round(250 + (220-250)*t))
	g := uint8(math.Round(100 + (240-100)*t))
	b := uint8(math.Round(0 + (240-0)*t))
	return pterm.NewRGB(r, g, b).Sprintf("%s", formatted)
}

func formatCommitTime(t time.Time) string {
	if t.IsZero() {
		return dimPlaceholder()
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
