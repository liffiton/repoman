package cmd

import (
	"fmt"
	"os"

	"github.com/liffiton/repoman/internal/api"
	"github.com/liffiton/repoman/internal/config"
	"github.com/liffiton/repoman/internal/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Repoman workspace in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintHeader("Initialize Current Directory")
		pterm.Println()

		if err := requireAuth(); err != nil {
			return err
		}

		// Check for existing workspace
		if root, err := config.FindWorkspaceRoot(); err == nil {
			curr, _ := os.Getwd()
			var msg string
			if root == curr {
				msg = "Current directory is already a Repoman workspace. Overwrite?"
			} else {
				ui.Warning.Printf("Found existing Repoman workspace at %s.\n", pterm.Bold.Sprint(root))
				msg = "Create a nested workspace here?"
			}

			result, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(msg).WithDefaultValue(false).Show()
			if !result {
				return nil
			}
		}

		client := api.NewClient(cfg.GetBaseURL(), cfg.APIKey)

		// 1. Select Course
		courses, err := client.GetCourses()
		if err != nil {
			return fmt.Errorf("failed to fetch courses: %w", err)
		}

		if len(courses) == 0 {
			return fmt.Errorf("no courses found")
		}

		var courseOptions []string
		courseMap := make(map[string]api.Course)
		for _, c := range courses {
			option := c.Name
			courseOptions = append(courseOptions, option)
			courseMap[option] = c
		}

		selectedCourseOption, err := pterm.DefaultInteractiveSelect.
			WithDefaultText("Select a course").
			WithOptions(courseOptions).
			Show()
		if err != nil {
			return err
		}
		selectedCourse := courseMap[selectedCourseOption]

		// 2. Select Assignment
		assignments, err := client.GetAssignments(selectedCourse.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch assignments: %w", err)
		}

		if len(assignments) == 0 {
			return fmt.Errorf("no assignments found for this course")
		}

		var assignmentOptions []string
		assignmentMap := make(map[string]api.Assignment)
		for _, a := range assignments {
			option := a.Name
			assignmentOptions = append(assignmentOptions, option)
			assignmentMap[option] = a
		}

		selectedAssignmentOption, err := pterm.DefaultInteractiveSelect.
			WithDefaultText("Select an assignment").
			WithOptions(assignmentOptions).
			Show()
		if err != nil {
			return err
		}
		selectedAssignment := assignmentMap[selectedAssignmentOption]

		// 3. Save Workspace Config
		wcfg := &config.WorkspaceConfig{
			CourseID:       selectedCourse.ID,
			CourseName:     selectedCourse.Name,
			AssignmentID:   selectedAssignment.ID,
			AssignmentName: selectedAssignment.Name,
		}

		if err := wcfg.SaveWorkspace(); err != nil {
			return fmt.Errorf("failed to save workspace config: %w", err)
		}

		ui.Success.Print("Current directory initialized ")
		fmt.Println("for " + pterm.Bold.Sprintf("%s - %s", selectedCourse.Name, selectedAssignment.Name))
		return nil
	},
}
