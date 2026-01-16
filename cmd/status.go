// Package cmd provides the entry point for the repoman CLI.
package cmd

import (
	"fmt"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/liffiton/repoman/internal/api"
	"github.com/liffiton/repoman/internal/config"
	"github.com/liffiton/repoman/internal/git"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all student repositories in the workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.APIKey == "" {
			return fmt.Errorf("not authenticated. Run 'repoman auth' first")
		}

		wcfg, err := config.LoadWorkspace()
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no workspace found. Run 'repoman init' first")
			}
			return fmt.Errorf("failed to load workspace: %w", err)
		}

		client := api.NewClient(cfg.GetBaseURL(), cfg.APIKey)
		repos, err := client.GetAssignmentRepos(wcfg.AssignmentID)
		if err != nil {
			return fmt.Errorf("failed to fetch repositories: %w", err)
		}

		fmt.Printf("Status for %s - %s\n\n", wcfg.CourseName, wcfg.AssignmentName)

		bar := progressbar.Default(int64(len(repos)), "Checking status")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "STUDENT/REPO\tBRANCH\tLAST COMMIT\tLOCAL STATUS\tSYNC STATE")

		var wg sync.WaitGroup
		results := make([]string, len(repos))

		sem := make(chan struct{}, 10) // Concurrency limit for status checks

		for i, repo := range repos {
			wg.Add(1)
			go func(i int, r api.Repo) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				localPath := r.Name
				if _, err := os.Stat(localPath); os.IsNotExist(err) {
					results[i] = fmt.Sprintf("%s\t-\t-\tMissing\t-", r.Name)
					_ = bar.Add(1)
					return
				}

				// Optional: Fetch to get accurate sync state.
				// Since it's 10-30 repos, we'll do it.
				_ = git.Fetch(localPath)

				branch, status, err := git.GetStatus(localPath)
				if err != nil {
					results[i] = fmt.Sprintf("%s\tERROR\t-\t%v\t-", r.Name, err)
					_ = bar.Add(1)
					return
				}

				syncState, _ := git.GetSyncState(localPath)
				lastCommit, _ := git.GetLastCommitTime(localPath)
				results[i] = fmt.Sprintf("%s\t%s\t%s\t%s\t%s", r.Name, branch, formatCommitTime(lastCommit), status, syncState)
				_ = bar.Add(1)
			}(i, repo)
		}

		wg.Wait()
		fmt.Println() // New line after progress bar

		for _, res := range results {
			_, _ = fmt.Fprintln(w, res)
		}
		_ = w.Flush()

		return nil
	},
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
