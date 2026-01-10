package cmd

import (
	"fmt"

	"github.com/liffiton/repoman/internal/update"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update repoman to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking for updates...")
		updated, err := update.CheckAndUpdate(version)
		if err != nil {
			return err
		}

		if updated {
			fmt.Println("Successfully updated to the latest version.")
		} else {
			fmt.Println("Repoman is already up to date.")
		}
		return nil
	},
}
