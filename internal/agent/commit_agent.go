package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-coders/gitchat/internal/git"
	"github.com/go-coders/gitchat/internal/llm"
	"github.com/go-coders/gitchat/pkg/apierrors"
	"github.com/go-coders/gitchat/pkg/utils"
)

type CommitAgent struct {
	git     git.Executor
	llm     *llm.Client
	display DisplayManager
	prompt  *CommitPrompt
	log     utils.Logger
}

func NewCommitAgent(git git.Executor, llm *llm.Client, display DisplayManager, log utils.Logger) *CommitAgent {
	return &CommitAgent{
		git:     git,
		llm:     llm,
		display: display,
		prompt:  NewCommitPrompt(),
		log:     log,
	}
}

// PrepareCommit prepares a commit by analyzing changes and generating suggestions
func (a *CommitAgent) PrepareCommit(ctx context.Context, opts CommitOptions) (staged []git.FileChange, unstaged []git.FileChange, response *CommitResponse, err error) {
	// check is git repo
	if !a.git.IsGitRepository(ctx) {
		return nil, nil, nil, apierrors.NewNotGitRepoError()
	}

	staged, unstaged, err = a.git.GetStatus(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Handle unstaged changes if needed
	if len(staged) == 0 && len(unstaged) > 0 && opts.AutoStage {
		a.display.ShowInfo("No staged changes found. Auto-staging all changes...")
		if err := a.git.StageAll(ctx); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to stage changes: %w", err)
		}
		staged, _, err = a.git.GetStatus(ctx)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get status: %w", err)
		}
	}

	// Generate suggestions if we have staged changes
	if len(staged) > 0 {
		a.display.StartSpinner("Analyzing changes and generating suggestions...")
		suggestions, err := a.generateSuggestions(ctx, staged)
		a.display.StopSpinner()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to generate suggestions: %w", err)
		}
		return staged, unstaged, suggestions, nil
	}

	return staged, unstaged, nil, nil
}

// Commit performs the actual commit operation
func (a *CommitAgent) Commit(ctx context.Context, message string) error {
	return a.git.Commit(ctx, message)
}

func (a *CommitAgent) generateSuggestions(ctx context.Context, files []git.FileChange) (*CommitResponse, error) {
	diff, err := a.git.GetDiff(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	prompt := a.prompt.GeneratePrompt(files, diff)
	response, err := a.llm.Complete(ctx, []llm.ConversationMessage{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	// Clean the response
	cleanedResponse := cleanJSONResponse(response)
	a.log.Debug("Cleaned LLM response: %s", cleanedResponse)

	var result CommitResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %w", err)
	}

	return &result, nil
}

func cleanJSONResponse(response string) string {
	// Remove markdown code blocks if present
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")

	// Trim spaces and newlines
	response = strings.TrimSpace(response)

	return response
}
