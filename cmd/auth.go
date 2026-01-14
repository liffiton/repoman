 // Package cmd provides the repoman CLI commands.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure authentication for the Repoman service",
	RunE: func(cmd *cobra.Command, args []string) error {
		var apiKey string
		fmt.Print("Your API key can be found in the Settings page of the Class Repo Manager web application.\n")
		fmt.Print("Enter API Key: ")
		if _, err := fmt.Scanln(&apiKey); err != nil && err.Error() != "unexpected newline" {
			return fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		var baseURL string
		fmt.Printf("Enter Base URL (default, if nothing entered: [%s]): ", cfg.GetBaseURL())
		if _, err := fmt.Scanln(&baseURL); err != nil && err.Error() != "unexpected newline" {
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

		fmt.Println("\nAuthentication configured successfully!")

		if result.KeyringUsed {
			fmt.Println("API Key: Saved securely in the system keyring.")
		} else {
			fmt.Printf("API Key: Saved in the config file (%s) because the system keyring was unavailable.\n", result.ConfigPath)
		}

		if result.FileWritten {
			fmt.Printf("Base URL: %s (saved in %s)\n", cfg.GetBaseURL(), result.ConfigPath)
		} else {
			fmt.Printf("Base URL: %s (using default, no config file created)\n", cfg.GetBaseURL())
		}

		return nil
	},
}
