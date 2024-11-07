package git

import "context"

// Logger defines the logging interface required by the git package
type Logger interface {
	Debug(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// FileChange represents a change to a file in git
type FileChange struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// Executor defines the interface for git operations
type Executor interface {
	// Basic git operations
	Execute(ctx context.Context, args ...string) (string, error)
	IsGitRepository(ctx context.Context) bool
	GetCurrentBranch(ctx context.Context) (string, error)

	// Status and staging operations
	GetStatus(ctx context.Context) (staged []FileChange, unstaged []FileChange, err error)
	StageAll(ctx context.Context) error
	StageFiles(ctx context.Context, files []string) error

	// Commit operations
	Commit(ctx context.Context, message string) error
	GetDiff(ctx context.Context, staged bool) (string, error)
}
