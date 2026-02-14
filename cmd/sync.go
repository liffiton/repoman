package cmd

import (
	"fmt"

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
		ctx, err := loadWorkspaceContext()
		if err != nil {
			return err
		}

		ui.PrintHeader(fmt.Sprintf("Syncing repositories for %s", pterm.Bold.Sprintf("%s - %s", ctx.Wcfg.CourseName, ctx.Wcfg.AssignmentName)))
		if ctx.OrigDir != ctx.Wcfg.Root {
			ui.Dim.Printf("Workspace: %s\n", ctx.Wcfg.Root)
		}
		pterm.Println()

		if len(ctx.Repos) == 0 {
			fmt.Println("No student repositories found for this assignment.")
			return nil
		}

		bar, _ := ui.Progressbar.WithTotal(len(ctx.Repos)).Start()

		manager := git.NewManager(6)
		var gitRepos []git.RepoInfo
		for _, r := range ctx.Repos {
			gitRepos = append(gitRepos, git.RepoInfo{
				Name:    r.Name,
				URL:     r.URL,
				Path:    r.Name, // Clone into current directory using the repo name
				UseHTTP: useHTTP,
			})
		}

		errs := manager.SyncAllCtx(cmd.Context(), gitRepos, func() {
			bar.Increment()
		})

		fmt.Println() // New line after progress bar

		successCount := 0
		for i, err := range errs {
			if err != nil {
				ui.Error.Printf("Error syncing %s: %v\n", ctx.Repos[i].Name, err)
			} else {
				successCount++
			}
		}

		fmt.Println(ui.Success.Sprint("Sync complete. ") + fmt.Sprintf("%d/%d repositories synced successfully.", successCount, len(ctx.Repos)))

		return nil
	},
}
