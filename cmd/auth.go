// Package cmd provides the repoman CLI commands.
package cmd

import (
	"fmt"
	"strings"

	"github.com/liffiton/repoman/internal/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure authentication for the Repoman service",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintHeader("Configure Authentication")
		pterm.Println()

		var apiKey string
		ui.Dim.Println("Your API key can be found in the Settings page of the Class Repo Manager web application.")
		apiKey, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Enter API Key").
			WithMask("*").
			Show()
		if err != nil {
			return fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		baseURL, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("Enter Base URL").
			WithDefaultValue(cfg.GetBaseURL()).
			Show()
		if err != nil {
			return fmt.Errorf("failed to read Base URL: %w", err)
		}
		baseURL = strings.TrimSpace(baseURL)

		cfg.APIKey = apiKey
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}

		result, err := cfg.Save()
		if err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		ui.Success.Println("\nAuthentication configured successfully!")

		if result.KeyringUsed {
			ui.Info.Println("API Key: Saved securely in the system keyring.")
		} else {
			ui.Info.Printf("API Key: Saved in the config file (%s) because the system keyring was unavailable.\n", result.ConfigPath)
		}

		if result.FileWritten {
			ui.Info.Printf("Base URL: %s (saved in %s)\n", cfg.GetBaseURL(), result.ConfigPath)
		} else {
			ui.Info.Printf("Base URL: %s (using default, no config file created)\n", cfg.GetBaseURL())
		}

		return nil
	},
}
