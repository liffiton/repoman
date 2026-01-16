package cmd

import (
	"fmt"
	"os"

	"github.com/liffiton/repoman/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "repoman",
	Short:   "Repoman is a CLI tool to manage Git repositories",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cmd.SilenceUsage = true // don't print usage for execution errors
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Root flags if any
}
