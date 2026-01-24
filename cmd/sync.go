package cmd

import (
	"fmt"
	"os"

	"github.com/liffiton/repoman/internal/api"
	"github.com/liffiton/repoman/internal/config"
	"github.com/liffiton/repoman/internal/git"
	"github.com/liffiton/repoman/internal/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var useHTTP bool

func init() {
	syncCmd.Flags().BoolVar(&useHTTP, "http", false, "Use HTTP instead of SSH for git operations")
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync student repositories for the current assignment",
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

		if len(repos) == 0 {
			fmt.Println("No student repositories found for this assignment.")
			return nil
		}

		ui.PrintHeader(fmt.Sprintf("Syncing %d repositories for ", len(repos)) + pterm.Bold.Sprintf("%s - %s", wcfg.CourseName, wcfg.AssignmentName))

		bar, _ := ui.Progressbar.WithTotal(len(repos)).Start()

		manager := git.NewManager(5)
		var gitRepos []git.RepoInfo
		for _, r := range repos {
			gitRepos = append(gitRepos, git.RepoInfo{
				URL:     r.URL,
				Path:    r.Name, // Clone into current directory using the repo name
				UseHTTP: useHTTP,
			})
		}

		errs := manager.SyncAll(gitRepos, func() {
			bar.Increment()
		})

		fmt.Println() // New line after progress bar

		successCount := 0
		for i, err := range errs {
			if err != nil {
				ui.Error.Printf("Error syncing %s: %v\n", repos[i].Name, err)
			} else {
				successCount++
			}
		}

		fmt.Println(ui.Success.Sprint("Sync complete. ") + fmt.Sprintf("%d/%d repositories synced successfully.", successCount, len(repos)))

		return nil
	},
}
