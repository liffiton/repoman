package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCourses(t *testing.T) {
	expectedCourses := []Course{
		{ID: "cs101", Name: "Intro to CS"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/courses" {
			t.Errorf("expected path /api/v1/courses, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedCourses)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	courses, err := client.GetCourses()
	if err != nil {
		t.Fatalf("GetCourses failed: %v", err)
	}
	if len(courses) != 1 || courses[0].ID != "cs101" {
		t.Errorf("unexpected courses: %+v", courses)
	}
}

func TestGetAssignments(t *testing.T) {
	expectedAssignments := []Assignment{
		{ID: "lab1", Name: "Lab 1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/courses/cs101/assignments" {
			t.Errorf("expected path /api/v1/courses/cs101/assignments, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedAssignments)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	assignments, err := client.GetAssignments("cs101")
	if err != nil {
		t.Fatalf("GetAssignments failed: %v", err)
	}
	if len(assignments) != 1 || assignments[0].ID != "lab1" {
		t.Errorf("unexpected assignments: %+v", assignments)
	}
}

func TestGetAssignmentRepos(t *testing.T) {
	expectedRepos := []Repo{
		{Name: "named-repo", URL: "https://github.com/user/named-repo"},
		{Name: "", URL: "https://github.com/user/unnamed-repo"},
		{Name: "unknown", URL: "git@github.com:user/unknown-repo.git"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/assignments/lab1/repos" {
			t.Errorf("expected path /api/v1/assignments/lab1/repos, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedRepos)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	repos, err := client.GetAssignmentRepos("lab1")
	if err != nil {
		t.Fatalf("GetAssignmentRepos failed: %v", err)
	}

	if len(repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(repos))
	}

	if repos[0].Name != "named-repo" {
		t.Errorf("expected named-repo, got %s", repos[0].Name)
	}
	if repos[1].Name != "unnamed-repo" {
		t.Errorf("expected unnamed-repo, got %s", repos[1].Name)
	}
	if repos[2].Name != "unknown-repo" {
		t.Errorf("expected unknown-repo, got %s", repos[2].Name)
	}
}
