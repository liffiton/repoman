package cmd

import (
	"testing"
	"time"

	"github.com/liffiton/repoman/internal/git"
)

func TestSortRepoStatusesByName(t *testing.T) {
	statuses := []git.RepoStatus{
		{Name: "Charlie", LastCommit: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "Alice", LastCommit: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
		{Name: "Bob", LastCommit: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	sortRepoStatuses(statuses, sortByName)

	want := []string{"Alice", "Bob", "Charlie"}
	for i, name := range want {
		if statuses[i].Name != name {
			t.Errorf("position %d: got %q, want %q", i, statuses[i].Name, name)
		}
	}
}

func TestSortRepoStatusesByDateOldestFirst(t *testing.T) {
	statuses := []git.RepoStatus{
		{Name: "Charlie", LastCommit: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Name: "Alice", LastCommit: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
		{Name: "Bob", LastCommit: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	sortRepoStatuses(statuses, sortByDate)

	want := []string{"Bob", "Charlie", "Alice"}
	for i, name := range want {
		if statuses[i].Name != name {
			t.Errorf("position %d: got %q, want %q", i, statuses[i].Name, name)
		}
	}
}

func TestSortRepoStatusesByDateKeepsBadReposLast(t *testing.T) {
	statuses := []git.RepoStatus{
		{Name: "Missing", Status: git.StatusMissing},
		{Name: "Older", LastCommit: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "Error", Status: git.StatusError},
		{Name: "Newer", LastCommit: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	sortRepoStatuses(statuses, sortByDate)

	want := []string{"Older", "Newer", "Error", "Missing"}
	for i, name := range want {
		if statuses[i].Name != name {
			t.Errorf("position %d: got %q, want %q", i, statuses[i].Name, name)
		}
	}
}

func TestSortRepoStatusesByDateUsesNameAsTieBreaker(t *testing.T) {
	sameTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	statuses := []git.RepoStatus{
		{Name: "Charlie", LastCommit: sameTime},
		{Name: "Alice", LastCommit: sameTime},
	}

	sortRepoStatuses(statuses, sortByDate)

	want := []string{"Alice", "Charlie"}
	for i, name := range want {
		if statuses[i].Name != name {
			t.Errorf("position %d: got %q, want %q", i, statuses[i].Name, name)
		}
	}
}

func TestValidateSortField(t *testing.T) {
	for _, field := range []string{sortByName, sortByDate} {
		if err := validateSortField(field); err != nil {
			t.Errorf("validateSortField(%q) returned unexpected error: %v", field, err)
		}
	}

	if err := validateSortField("last-commit"); err == nil {
		t.Error("validateSortField(\"last-commit\") returned nil, want error")
	}
}
