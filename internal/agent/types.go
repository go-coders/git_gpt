package agent

import (
	"context"
	"io"

	"github.com/go-coders/git_gpt/internal/common"
)

// Core interfaces for dependencies
type (
	Logger interface {
		Debug(format string, args ...interface{})
		Error(format string, args ...interface{})
	}

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

	LLMClient interface {
		Chat(ctx context.Context, content string) (string, error)
		SetSystemMessage(message string)
		ClearHistory()
	}

	GitExecutor interface {
		Execute(ctx context.Context, args ...string) (string, error)
		IsGitRepository(ctx context.Context) bool
		GetStatus(ctx context.Context) (staged []common.FileChange, unstaged []common.FileChange, err error)
		StageAll(ctx context.Context) error
		StageFiles(ctx context.Context, files []string) error
		Commit(ctx context.Context, message string) error
		GetDiff(ctx context.Context, staged bool) (string, error)
	}

	InputReader interface {
		ReadString(delim byte) (string, error)
	}
)

// Command types
const (
	CommandTypeQuery  = "query"
	CommandTypeModify = "modify"
)

// Core data structures
type (
	Command struct {
		Type    string   `json:"type"`
		Action  string   `json:"action"`
		Args    []string `json:"args"`
		Purpose string   `json:"purpose"`
		Impact  string   `json:"impact,omitempty"`
	}

	Response struct {
		Type        string    `json:"type"`
		CommandType string    `json:"commandType,omitempty"`
		Content     string    `json:"content"`
		Commands    []Command `json:"commands"`
		Reason      string    `json:"reason"`
	}

	CommandResult struct {
		Command Command
		Output  string
		Error   error
	}

	CommitResponse struct {
		Summary     string             `json:"summary"`
		Suggestions []CommitSuggestion `json:"suggestions"`
	}

	CommitSuggestion struct {
		Message     string `json:"message"`
		Description string `json:"description,omitempty"`
	}

	CommitOptions struct {
		AutoStage bool
	}

	// Agent configuration
	AgentConfig struct {
		Git     GitExecutor
		LLM     LLMClient
		Display DisplayManager
		Logger  Logger
		Reader  io.Reader
	}
)

// Validator interface for command validation
