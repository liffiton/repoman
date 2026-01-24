// Package ui provides common UI components and styles for the Repoman CLI.
package ui

import (
	"github.com/pterm/pterm"
)

var (
	// Success is the style for success messages
	Success = pterm.NewRGB(80, 180, 40)

	// Error is the style for error messages
	Error = pterm.Error

	// Info is the style for info messages
	Info = pterm.NewRGB(80, 180, 200)

	// Normal is the style for plain text
	Normal = pterm.DefaultBasicText

	// Dim is the style for secondary/dimmed text
	Dim = pterm.FgGray.ToStyle()

	// Progressbar sets up a styled progress bar
	Progressbar = pterm.DefaultProgressbar.WithBarStyle(pterm.FgGray.ToStyle()).WithBarFiller(pterm.Gray("."))
)

// PrintHeader prints a header at the start of the program
func PrintHeader(title string) {
	RepomanTitle := pterm.NewRGB(60, 140, 250)
	RepomanTitle.Print("Repoman: ")
	Normal.Println(title)
}
