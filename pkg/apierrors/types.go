package apierrors

import (
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// User-facing errors that require specific handling
	ErrTokenLimitExceeded ErrorType = "token_limit_exceeded"
	ErrInvalidAPIKey      ErrorType = "invalid_api_key"
	ErrInvailidModel      ErrorType = "invalid_model"
	ErrGitNotInitialized  ErrorType = "git_not_initialized"
)

// AppError represents an application error with context
type AppError struct {
	Type     ErrorType
	Message  string
	Metadata map[string]interface{}
}

func (e *AppError) Error() string {

	return e.Message
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:     errType,
		Message:  message,
		Metadata: make(map[string]interface{}),
	}
}

// Error constructors for common cases
func NewTokenLimitError(current, max int) *AppError {
	return &AppError{
		Type:    ErrTokenLimitExceeded,
		Message: fmt.Sprintf("token limit exceeded: current %d, max %d", current, max),
		Metadata: map[string]interface{}{
			"current_tokens": current,
			"max_tokens":     max,
		},
	}
}

func NewNotGitRepoError() *AppError {
	return &AppError{
		Type:    ErrGitNotInitialized,
		Message: "Current  directory is not a git repository, please run cd <path> to change to a git repository",
	}
}

func NewAPIKeyOrUrlError(err error) *AppError {
	return &AppError{
		Type:    ErrInvalidAPIKey,
		Message: fmt.Sprintf("invalid API key or base URL: %v", err),
	}
}

func NewApiKeyError() *AppError {
	return &AppError{
		Type:    ErrInvalidAPIKey,
		Message: "API key is invalid",
	}
}

func NewInvalidModelError() *AppError {
	return &AppError{
		Type:    ErrInvailidModel,
		Message: "Model is invalid",
	}
}
