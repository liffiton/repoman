package cmd

import (
	"fmt"
	"os"

	"github.com/liffiton/repoman/internal/api"
	"github.com/liffiton/repoman/internal/config"
)

// requireAuth ensures the user is authenticated.
func requireAuth() error {
	if cfg.APIKey == "" {
		return fmt.Errorf("not authenticated. Run 'repoman auth' first")
	}
	return nil
}

// workspaceContext holds the context for a workspace-related command.
type workspaceContext struct {
	Client  *api.Client
	Wcfg    *config.WorkspaceConfig
	OrigDir string
	Repos   []api.Repo
}

// loadWorkspaceContext loads the workspace configuration, changes to the root directory,
// and fetches the assignment repositories.
func loadWorkspaceContext() (*workspaceContext, error) {
	if err := requireAuth(); err != nil {
		return nil, err
	}

	wcfg, err := config.LoadWorkspace()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no workspace found. Run 'repoman init' first")
		}
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(wcfg.Root); err != nil {
		return nil, fmt.Errorf("failed to change to workspace root: %w", err)
	}

	client := api.NewClient(cfg.GetBaseURL(), cfg.APIKey)
	repos, err := client.GetAssignmentRepos(wcfg.AssignmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	return &workspaceContext{
		Client:  client,
		Wcfg:    wcfg,
		Repos:   repos,
		OrigDir: origDir,
	}, nil
}
