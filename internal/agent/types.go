// internal/agent/types.go
package agent

// DisplayManager handles UI interactions
type DisplayManager interface {
	ShowInfo(message string)
	ShowError(message string)
	ShowSuccess(message string)
	ShowWarning(message string)
	StartSpinner(message string)
	StopSpinner()
	ShowCommand(command string)
}

// CommitResult represents the result of a commit operation
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
