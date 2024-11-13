package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-coders/git_gpt/internal/common"
)

type CommitAgent struct {
	*BaseAgent
}

func NewCommitAgent(config AgentConfig) (*CommitAgent, error) {
	base, err := NewBaseAgent(config)
	if err != nil {
		return nil, err
	}

	return &CommitAgent{
		BaseAgent: base,
	}, nil
}

// HandleCommit manages the entire commit process
func (a *CommitAgent) HandleCommit(ctx context.Context) error {
	status, err := a.prepareCommit(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare commit: %w", err)
	}

	// Handle different commit scenarios
	switch {
	case a.hasNoChanges(status):
		a.display.ShowInfo("No changes to commit")
		return nil

	case a.hasOnlyUnstagedChanges(status):
		shoudCommit, err := a.handleUnstagedChanges(ctx, status.unstaged)

		if err != nil {
			return fmt.Errorf("failed to handle unstaged changes: %w", err)
		}

		if !shoudCommit {
			return nil
		}

		return a.HandleCommit(ctx)

	case len(status.staged) > 0:
		return a.processStagedChanges(ctx, status)

	default:
		return fmt.Errorf("unexpected repository state")
	}
}

type commitStatus struct {
	staged      []common.FileChange
	unstaged    []common.FileChange
	suggestions *CommitResponse
}

func (a *CommitAgent) prepareCommit(ctx context.Context) (*commitStatus, error) {
	staged, unstaged, err := a.git.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var suggestions *CommitResponse
	if len(staged) > 0 {
		suggestions, err = a.generateCommitSuggestions(ctx, staged)
		if err != nil {
			return nil, err
		}
	}

	return &commitStatus{
		staged:      staged,
		unstaged:    unstaged,
		suggestions: suggestions,
	}, nil
}

func (a *CommitAgent) hasNoChanges(status *commitStatus) bool {
	return len(status.staged) == 0 && len(status.unstaged) == 0
}

func (a *CommitAgent) hasOnlyUnstagedChanges(status *commitStatus) bool {
	return len(status.unstaged) > 0 && len(status.staged) == 0
}

func (a *CommitAgent) handleUnstagedChanges(ctx context.Context, unstaged []common.FileChange) (bool, error) {
	modified, untracked := a.categorizeChanges(unstaged)

	// Display changes
	a.displayUnstagedChanges(modified, untracked)

	// Prompt for staging
	confirmed, err := a.promptForConfirmation("\nWould you like to stage all changes? (y/n): ")
	if err != nil {
		return false, fmt.Errorf("failed to prompt for confirmation: %w", err)
	}
	fmt.Println(confirmed, err, "confirmed")
	if !confirmed {
		a.display.ShowInfo("Commit cancelled")
		return false, nil
	}

	// Stage changes
	if err := a.git.StageAll(ctx); err != nil {
		return false, fmt.Errorf("failed to stage changes: %w", err)
	}

	a.display.ShowSuccess("All changes staged successfully")
	return true, nil
}

func (a *CommitAgent) processStagedChanges(ctx context.Context, status *commitStatus) error {
	if status.suggestions == nil || len(status.suggestions.Suggestions) == 0 {
		a.display.ShowInfo("No commit suggestions found")
		return nil
	}

	a.displayStagedChanges(status.staged)
	a.displayCommitSuggestions(status.suggestions)

	message, regenerate, err := a.getCommitMessage(status.suggestions.Suggestions)
	if err != nil {
		return err
	}

	if regenerate {
		return a.HandleCommit(ctx)
	}

	if message == "" {
		a.display.ShowInfo("Commit cancelled")
		return nil
	}

	if err := a.git.Commit(ctx, message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	a.display.ShowSuccess(fmt.Sprintf("Changes committed successfully with message: %s", message))
	return nil
}

func (a *CommitAgent) generateCommitSuggestions(ctx context.Context, files []common.FileChange) (*CommitResponse, error) {
	a.display.StartSpinner("Analyzing changes and generating suggestions...")
	defer a.display.StopSpinner()

	diff, err := a.git.GetDiff(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	prompt, err := a.prompts.GetCommitPrompt(files, diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commit prompt: %w", err)
	}

	response, err := a.llm.Chat(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	cleanedResponse := cleanJSONResponse(response)
	a.logger.Debug("Cleaned LLM response: %s", cleanedResponse)

	var result CommitResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %w", err)
	}

	return &result, nil
}

func (a *CommitAgent) categorizeChanges(changes []common.FileChange) (modified, untracked []string) {
	modified = make([]string, 0)
	untracked = make([]string, 0)

	for _, change := range changes {
		if change.Status == "untracked" {
			untracked = append(untracked, change.Path)
		} else {
			modified = append(modified, change.Path)
		}
	}
	return modified, untracked
}

func (a *CommitAgent) displayUnstagedChanges(modified, untracked []string) {
	if len(modified) > 0 {
		a.display.ShowSection("Modified Files", "", map[string]string{"icon": "📝"})
		for _, path := range modified {
			a.display.ShowInfo(fmt.Sprintf("  %s", path))
		}
	}

	if len(untracked) > 0 {
		a.display.ShowSection("Untracked Files", "", map[string]string{"icon": "❓"})
		for _, path := range untracked {
			a.display.ShowInfo(fmt.Sprintf("  %s", path))
		}
	}
}

func (a *CommitAgent) displayStagedChanges(changes []common.FileChange) {
	items := make([][2]string, 0, len(changes))
	for _, change := range changes {
		items = append(items, [2]string{
			fmt.Sprintf("%s %s", getStatusSymbol(change.Status), change.Path),
			fmt.Sprintf("(%d+/%d-)", change.Additions, change.Deletions),
		})
	}

	a.display.ShowSection("Staged Files", "", map[string]string{
		"icon":    "📄",
		"divider": "------------------------",
	})
	a.display.ShowNumberedList(items)
}

func (a *CommitAgent) displayCommitSuggestions(suggestions *CommitResponse) {
	a.display.ShowSection("Change Summary", suggestions.Summary, map[string]string{
		"icon":    "📝",
		"divider": "------------------------",
	})

	items := make([][2]string, 0, len(suggestions.Suggestions))
	for _, suggestion := range suggestions.Suggestions {
		items = append(items, [2]string{
			suggestion.Message,
			suggestion.Description,
		})
	}

	a.display.ShowSection("Suggested Commit Messages", "", map[string]string{"icon": "💡"})
	a.display.ShowNumberedList(items)
}

func (a *CommitAgent) getCommitMessage(suggestions []CommitSuggestion) (string, bool, error) {
	fmt.Print("\nSelect a message (1-3), 'r' to regenerate, 'c' to cancel, or 'm' for manual input: ")
	input, err := a.reader.ReadString('\n')
	if err != nil {
		return "", false, fmt.Errorf("failed to read input: %w", err)
	}

	return a.processCommitMessageInput(strings.TrimSpace(input), suggestions)
}

func (a *CommitAgent) processCommitMessageInput(input string, suggestions []CommitSuggestion) (string, bool, error) {
	switch input {
	case "c":
		return "", false, nil
	case "m":
		return a.getManualCommitMessage()
	case "r":
		return "", true, nil
	case "":
		return "", false, nil
	default:
		return a.processNumberedSelection(input, suggestions)
	}
}

func (a *CommitAgent) getManualCommitMessage() (string, bool, error) {
	fmt.Print("Enter your commit message: ")
	message, err := a.reader.ReadString('\n')
	if err != nil {
		return "", false, fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(message), false, nil
}

func (a *CommitAgent) processNumberedSelection(input string, suggestions []CommitSuggestion) (string, bool, error) {
	selection := 0
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
		return "", false, fmt.Errorf("invalid selection")
	}

	if selection < 1 || selection > len(suggestions) {
		return "", false, fmt.Errorf("invalid selection: must be between 1 and %d", len(suggestions))
	}

	return suggestions[selection-1].Message, false, nil
}

func getStatusSymbol(status string) string {
	symbols := map[string]string{
		"modified":  "📝",
		"added":     "➕",
		"deleted":   "➖",
		"renamed":   "📋",
		"copied":    "📑",
		"untracked": "❓",
	}

	if symbol, ok := symbols[status]; ok {
		return symbol
	}
	return "•"
}
