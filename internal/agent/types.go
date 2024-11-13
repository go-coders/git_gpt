// agent/types.go
package agent

import (
	"context"

	"github.com/go-coders/git_gpt/internal/git"
)

// Core interfaces that the agent package depends on
type (
	// Logger defines logging capabilities required by agents
	Logger interface {
		Debug(format string, args ...interface{})
		Error(format string, args ...interface{})
	}

	// DisplayManager handles UI interactions
	DisplayManager interface {
		ShowInfo(message string)
		ShowError(message string)
		ShowSuccess(message string)
		ShowWarning(message string)
		StartSpinner(message string)
		StopSpinner()
		ShowCommand(command string)
		ShowSection(title, content string, opts map[string]string)
		ShowNumberedList(items [][2]string)
	}

	// LLMClient defines the language model capabilities
	LLMClient interface {
		Chat(ctx context.Context, content string) (string, error)
		SetSystemMessage(message string)
		ClearHistory()
	}

	// GitExecutor defines required git operations
	GitExecutor interface {
		Execute(ctx context.Context, args ...string) (string, error)
		IsGitRepository(ctx context.Context) bool
		GetStatus(ctx context.Context) (staged []git.FileChange, unstaged []git.FileChange, err error)
		StageAll(ctx context.Context) error
		StageFiles(ctx context.Context, files []string) error
		Commit(ctx context.Context, message string) error
		GetDiff(ctx context.Context, staged bool) (string, error)
	}
)

// Section represents a display section with formatting
type Section struct {
	Title    string
	Content  string
	Icon     string
	Divider  string
	Numbered bool
}

// ListItem represents an item in a numbered list with description
type ListItem struct {
	Content     string
	Description string
}

// Response represents the LLM response structure
type Response struct {
	Type     string       `json:"type"`
	Content  string       `json:"content"`
	Commands []GitCommand `json:"commands"`
	Reason   string       `json:"reason"`
}

// GitCommand represents a git command with its arguments and purpose
type GitCommand struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Purpose string   `json:"purpose"`
}

// CommandResult represents the result of executing a git command
type CommandResult struct {
	Command GitCommand
	Output  string
	Error   error
}

// CommitResponse represents the response for commit suggestions
type CommitResponse struct {
	Summary     string             `json:"summary"`
	Suggestions []CommitSuggestion `json:"suggestions"`
}

// CommitSuggestion represents a suggested commit message
type CommitSuggestion struct {
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// CommitOptions contains options for the commit operation
type CommitOptions struct {
	AutoStage bool
}

// Agent represents the main chat agent
