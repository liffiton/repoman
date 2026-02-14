package git

import (
	"context"
	"os"
	"sync"
	"time"
)

// RepoInfo contains information about a repository to be managed.
type RepoInfo struct {
	Name    string
	URL     string
	Path    string
	UseHTTP bool
}

// RepoStatus contains the status of a repository.
type RepoStatus struct {
	Error      error
	LastCommit time.Time
	Name       string
	Branch     string
	Status     string
	SyncState  string
}

const (
	// StatusMissing indicates the repository directory does not exist.
	StatusMissing = "Missing"
	// StatusError indicates an error occurred while checking the repository status.
	StatusError = "Error"
	// StateUnknown indicates the sync state of the repository is unknown.
	StateUnknown = "Unknown"
	// StateStale indicates the repository is behind the remote.
	StateStale = "Stale"
	// StateSynced indicates the repository is up to date with the remote.
	StateSynced = "Synced"
)

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
	return m.SyncAllCtx(context.Background(), repos, progress)
}

// SyncAllCtx syncs all provided repositories concurrently.
// Uses the provided context for timeout/cancellation control.
// If progress is not nil, it is called after each repository is synced.
func (m *Manager) SyncAllCtx(ctx context.Context, repos []RepoInfo, progress func()) []error {
	worker := func(ctx context.Context, r RepoInfo) error {
		return SyncCtx(ctx, r.URL, r.Path, r.UseHTTP)
	}
	return concurrentMap(ctx, m.concurrency, repos, worker, progress)
}

// StatusAll fetches status for all provided repositories concurrently.
// If progress is not nil, it is called after each repository's status is checked.
func (m *Manager) StatusAll(repos []RepoInfo, fetch bool, progress func()) []RepoStatus {
	return m.StatusAllCtx(context.Background(), repos, fetch, progress)
}

// StatusAllCtx fetches status for all provided repositories concurrently.
// Uses the provided context for timeout/cancellation control.
// If progress is not nil, it is called after each repository's status is checked.
func (m *Manager) StatusAllCtx(ctx context.Context, repos []RepoInfo, fetch bool, progress func()) []RepoStatus {
	worker := func(ctx context.Context, r RepoInfo) RepoStatus {
		return fetchStatusWithCtx(ctx, r, fetch)
	}
	return concurrentMap(ctx, m.concurrency, repos, worker, progress)
}

func fetchStatusWithCtx(ctx context.Context, r RepoInfo, fetch bool) RepoStatus {
	status := RepoStatus{Name: r.Name}

	if _, err := os.Stat(r.Path); os.IsNotExist(err) {
		status.Status = StatusMissing
		return status
	}

	var fetchErr error
	if fetch {
		fetchCtx, fetchCancel := context.WithTimeout(ctx, defaultPullTimeout)
		defer fetchCancel()
		fetchErr = FetchCtx(fetchCtx, r.Path)
	}

	branch, repoSummary, err := GetStatusCtx(ctx, r.Path)
	status.Branch = branch
	if err != nil {
		status.Status = StatusError
		status.Error = err
	} else {
		status.Status = repoSummary
	}

	syncState, syncErr := GetSyncStateCtx(ctx, r.Path)
	if syncErr != nil {
		status.SyncState = StateUnknown
		if status.Error == nil {
			status.Error = syncErr
		}
	} else {
		status.SyncState = syncState
		if fetchErr != nil {
			status.SyncState += " (" + StateStale + ")"
		}
	}

	lastCommit, err := GetLastCommitTimeCtx(ctx, r.Path)
	status.LastCommit = lastCommit
	if err != nil && status.Error == nil {
		status.Error = err
	}

	return status
}

// concurrentMap transforms a slice of T into a slice of R concurrently using a worker pool.
// It respects context cancellation and will stop early if the context is canceled.
func concurrentMap[T any, R any](ctx context.Context, concurrency int, items []T, worker func(context.Context, T) R, progress func()) []R {
	results := make([]R, len(items))
	if len(items) == 0 {
		return results
	}

	type task struct {
		item  T
		index int
	}

	tasks := make(chan task, len(items))
	for i, item := range items {
		tasks <- task{item, i}
	}
	close(tasks)

	var wg sync.WaitGroup
	var mu sync.Mutex
	numWorkers := concurrency
	if numWorkers > len(items) {
		numWorkers = len(items)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-tasks:
					if !ok {
						return
					}
					res := worker(ctx, t.item)
					results[t.index] = res
					if progress != nil {
						mu.Lock()
						progress()
						mu.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()
	return results
}
