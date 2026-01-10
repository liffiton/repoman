package git

import (
	"sync"
)

// RepoInfo contains information about a repository to be managed.
type RepoInfo struct {
	URL     string
	Path    string
	UseHTTP bool
}

// Manager handles concurrent git operations.
type Manager struct {
	concurrency int
}

// NewManager creates a new Manager with the specified concurrency limit.
func NewManager(concurrency int) *Manager {
	if concurrency <= 0 {
		concurrency = 5
	}
	return &Manager{concurrency: concurrency}
}

// SyncAll syncs all provided repositories concurrently.
// If progress is not nil, it is called after each repository is synced.
func (m *Manager) SyncAll(repos []RepoInfo, progress func()) []error {
	var wg sync.WaitGroup
	errs := make([]error, len(repos))
	sem := make(chan struct{}, m.concurrency)

	for i, repo := range repos {
		wg.Add(1)
		go func(i int, r RepoInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			errs[i] = Sync(r.URL, r.Path, r.UseHTTP)
			if progress != nil {
				progress()
			}
		}(i, repo)
	}

	wg.Wait()
	return errs
}
