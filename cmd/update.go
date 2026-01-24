package cmd

import (
	"fmt"

	"github.com/liffiton/repoman/internal/ui"
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
		ui.PrintHeader("Checking for updates...")
		fmt.Println()

		updated, err := update.CheckAndUpdate(version)
		if err != nil {
			return err
		}

		if updated {
			fmt.Println()
			ui.Success.Print("Successfully updated ")
			fmt.Println("to the latest version.")
		} else {
			fmt.Println("Repoman is already up to date.")
		}
		return nil
	},
}
