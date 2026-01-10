// Package api provides a client for the Repoman web application.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/liffiton/repoman/internal/git"
)

// Course represents a course in the web application.
type Course struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Assignment represents an assignment in a course.
type Assignment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Repo represents a git repository for an assignment.
type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Client is a client for the Repoman web application.
type Client struct {
	baseURL string
	apiKey  string
}

// NewClient creates a new API client.
func NewClient(baseURL, apiKey string) *Client {
	// Ensure baseURL doesn't end with a slash for consistent path joining
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (c *Client) doRequest(method, path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1%s", c.baseURL, path)
	req, err := http.NewRequest(method, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unauthorized: invalid API key")
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}

// GetCourses fetches the list of courses.
func (c *Client) GetCourses() ([]Course, error) {
	resp, err := c.doRequest("GET", "/courses")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var courses []Course
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode courses: %w", err)
	}
	return courses, nil
}

// GetAssignments fetches the list of assignments for a course.
func (c *Client) GetAssignments(courseID string) ([]Assignment, error) {
	path := fmt.Sprintf("/courses/%s/assignments", courseID)
	resp, err := c.doRequest("GET", path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var assignments []Assignment
	if err := json.NewDecoder(resp.Body).Decode(&assignments); err != nil {
		return nil, fmt.Errorf("failed to decode assignments: %w", err)
	}
	return assignments, nil
}

// GetAssignmentRepos fetches the list of repositories for an assignment.
func (c *Client) GetAssignmentRepos(assignmentID string) ([]Repo, error) {
	path := fmt.Sprintf("/assignments/%s/repos", assignmentID)
	resp, err := c.doRequest("GET", path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repos: %w", err)
	}

	// Post-process to ensure names are populated
	for i := range repos {
		if repos[i].Name == "" || repos[i].Name == "unknown" {
			repos[i].Name = git.ExtractRepoName(repos[i].URL)
		}
	}

	return repos, nil
}
