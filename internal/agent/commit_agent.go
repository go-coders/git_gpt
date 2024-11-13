// internal/agent/commit_agent.go
package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-coders/gitchat/internal/git"
)

type CommitAgent struct {
	git     GitExecutor
	llm     LLMClient
	display DisplayManager
	prompt  *PromptManager
	log     Logger
	reader  *bufio.Reader
}

func NewCommitAgent(git GitExecutor, llm LLMClient, display DisplayManager, log Logger) (*CommitAgent, error) {

	prompts, err := NewPromptManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt manager: %w", err)
	}

	return &CommitAgent{
		git:     git,
		llm:     llm,
		display: display,
		prompt:  prompts,
		log:     log,
		reader:  bufio.NewReader(os.Stdin),
	}, nil
}

// HandleCommit manages the entire commit process
func (a *CommitAgent) HandleCommit(ctx context.Context) error {
	for {
		staged, unstaged, suggestions, err := a.PrepareCommit(ctx, CommitOptions{AutoStage: false})
		if err != nil {
			return fmt.Errorf("failed to prepare commit: %w", err)
		}

		// Check if there are any changes to commit
		if len(staged) == 0 && len(unstaged) == 0 {
			a.display.ShowInfo("No changes to commit")
			return nil
		}

		// Handle unstaged changes
		if len(unstaged) > 0 && len(staged) == 0 {
			cont, err := a.handleUnstagedChanges(ctx, unstaged)
			if err != nil {
				return fmt.Errorf("failed to handle unstaged changes: %w", err)
			}
			if !cont {
				return nil
			}
			continue
		}

		// Check for valid suggestions
		if suggestions == nil || len(suggestions.Suggestions) == 0 {
			a.display.ShowInfo("No commit suggestions found")
			return nil
		}

		// Display changes and get commit message
		if len(staged) > 0 {
			a.displayStagedChanges(staged)
		}

		a.displayCommitSuggestions(suggestions)
		message, regenerate, err := a.getCommitMessage(suggestions.Suggestions)
		if err != nil {
			return err
		}
		if regenerate {
			continue
		}
		if message == "" {
			a.display.ShowInfo("Commit cancelled")
			return nil
		}

		// Perform the commit
		if err := a.Commit(ctx, message); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}

		a.display.ShowSuccess(fmt.Sprintf("Changes committed successfully with message: %s", message))
		return nil
	}
}

func (a *CommitAgent) handleUnstagedChanges(ctx context.Context, unstaged []git.FileChange) (bool, error) {
	modified, untracked := a.categorizeChanges(unstaged)

	// Display modified files
	if len(modified) > 0 {
		fmt.Println("\nModified files:")
		for _, path := range modified {
			fmt.Printf("  %s\n", path)
		}
	}

	// Display untracked files
	if len(untracked) > 0 {
		fmt.Println("\nUntracked files:")
		for _, path := range untracked {
			fmt.Printf("  %s\n", path)
		}
	}

	// Prompt for staging
	if cont, err := a.promptForStaging(); err != nil {
		return false, err
	} else if !cont {
		return false, nil
	}

	// Stage all changes
	if err := a.git.StageAll(ctx); err != nil {
		return false, fmt.Errorf("failed to stage changes: %w", err)
	}

	a.display.ShowSuccess("All changes staged successfully")
	return true, nil
}

func (a *CommitAgent) categorizeChanges(changes []git.FileChange) (modified, untracked []string) {
	for _, change := range changes {
		if change.Status == "untracked" {
			untracked = append(untracked, change.Path)
		} else {
			modified = append(modified, change.Path)
		}
	}
	return modified, untracked
}

func (a *CommitAgent) promptForStaging() (bool, error) {
	fmt.Print("\nWould you like to stage all changes? (y/n): ")
	input, err := a.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	if strings.TrimSpace(input) != "y" {
		a.display.ShowInfo("Commit cancelled")
		return false, nil
	}
	return true, nil
}

func (a *CommitAgent) displayStagedChanges(changes []git.FileChange) {
	fmt.Println("\n📄 Staged Files:")
	fmt.Println("------------------------")
	for _, change := range changes {
		fmt.Printf("%s %s (%d+/%d-)\n",
			getStatusSymbol(change.Status),
			change.Path,
			change.Additions,
			change.Deletions,
		)
	}
}

func (a *CommitAgent) displayCommitSuggestions(suggestions *CommitResponse) {
	a.display.ShowSection("Change Summary", suggestions.Summary, map[string]string{
		"icon":    "📝",
		"divider": "------------------------",
	})

	// Display suggestions section
	a.display.ShowSection("Suggested Commit Messages", "", map[string]string{
		"icon": "💡",
	})

	// Convert suggestions to list items
	items := make([][2]string, 0, len(suggestions.Suggestions))
	for _, suggestion := range suggestions.Suggestions {
		items = append(items, [2]string{
			suggestion.Message,
			suggestion.Description,
		})
	}
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
		return "", false, fmt.Errorf("invalid selection")
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

func (a *CommitAgent) PrepareCommit(ctx context.Context, opts CommitOptions) (staged []git.FileChange, unstaged []git.FileChange, response *CommitResponse, err error) {
	staged, unstaged, err = a.git.GetStatus(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get status: %w", err)
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

// Commit performs the commit operation
func (a *CommitAgent) Commit(ctx context.Context, message string) error {
	return a.git.Commit(ctx, message)
}

func (a *CommitAgent) generateSuggestions(ctx context.Context, files []git.FileChange) (*CommitResponse, error) {
	diff, err := a.git.GetDiff(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	prompt, err := a.prompt.GetCommitPrompt(files, diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commit prompt: %w", err)
	}

	response, err := a.llm.Chat(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	cleanedResponse := cleanJSONResponse(response)
	a.log.Debug("Cleaned LLM response: %s", cleanedResponse)

	var result CommitResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %w", err)
	}

	return &result, nil
}
func cleanJSONResponse(response string) string {
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")

	response = strings.TrimSpace(response)

	return response
}
