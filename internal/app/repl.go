package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-coders/gitchat/internal/agent"
	"github.com/go-coders/gitchat/internal/display"
	"github.com/go-coders/gitchat/internal/git"
	"github.com/go-coders/gitchat/internal/version"
)

type REPL struct {
	app    *Application
	reader *bufio.Reader
}

func NewREPL(app *Application) *REPL {
	return &REPL{
		app:    app,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (r *REPL) Start(ctx context.Context) error {
	for {
		if err := r.showPrompt(ctx); err != nil {
			return err
		}

		input, err := r.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("input error: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			return nil
		}

		if err := r.handleInput(ctx, input); err != nil {
			r.app.HandleErr(err)
		}
	}
}

func (r *REPL) showPrompt(ctx context.Context) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	branch := "no git"
	if r.app.gitClient.IsGitRepository(ctx) {
		if b, err := r.app.gitClient.GetCurrentBranch(ctx); err == nil {
			branch = b
		}
	}

	r.app.display.ShowPrompt(pwd, branch)
	return nil
}

func (r *REPL) handleInput(ctx context.Context, input string) error {
	switch {
	case input == "version":
		r.app.display.ShowInfo(version.GetVersionInfo())
		return nil
	case input == "config":
		wizard := NewConfigWizard(r.app.config, r.app.logger)
		if err := wizard.Run(); err != nil {
			return err
		}
		if err := r.app.Reload(); err != nil {
			return fmt.Errorf("failed to reload application: %w", err)
		}
		r.app.display.ShowSuccess("Configuration updated and reloaded successfully")
		return nil
	case input == "commit":
		return r.handleCommit(ctx)
	case strings.HasPrefix(input, "cd"):
		return r.handleChangeDirectory(input)
	// hanele empty input
	case strings.TrimSpace(input) == "":
		return nil
	default:
		return r.app.agent.Chat(ctx, input)
	}
}

func (r *REPL) handleUnstage(ctx context.Context, unstaged []git.FileChange) (bool, error) {
	modified := make([]string, 0)
	untracked := make([]string, 0)
	for _, change := range unstaged {
		if change.Status == "untracked" {
			untracked = append(untracked, change.Path)
		} else {
			modified = append(modified, change.Path)
		}
	}
	// Show all modified files
	if len(modified) > 0 {
		fmt.Println("\nModified files:")
		for _, path := range modified {
			fmt.Printf("  %s\n", path)
		}
	}

	// Show all untracked files
	if len(untracked) > 0 {
		fmt.Println("\nUntracked files:")
		for _, path := range untracked {
			fmt.Printf("  %s\n", path)
		}
	}

	fmt.Print("\nWould you like to stage all changes? (y/n): ")
	input, err := r.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	if strings.TrimSpace(input) != "y" {
		r.app.display.ShowInfo("Commit cancelled")
		return false, nil
	}

	if err := r.app.gitClient.StageAll(ctx); err != nil {
		return false, fmt.Errorf("failed to stage all changes: %w", err)
	}

	r.app.display.ShowSuccess("All changes staged successfully")
	return true, nil

}

func (r *REPL) handleCommit(ctx context.Context) error {

	for {
		staged, unstaged, suggestions, err := r.app.commitAgent.PrepareCommit(ctx, agent.CommitOptions{
			AutoStage: false,
		})
		if err != nil {
			return fmt.Errorf("failed to prepare commit: %w", err)
		}
		if len(staged) == 0 && len(unstaged) == 0 {
			r.app.display.ShowInfo("No changes to commit")
			return nil
		}
		if len(unstaged) > 0 && len(staged) == 0 {
			conti, err := r.handleUnstage(ctx, unstaged)
			if err != nil {
				return fmt.Errorf("failed to stage all changes: %w", err)
			}
			if !conti {
				return nil
			}
			continue
		}

		if suggestions == nil || len(suggestions.Suggestions) == 0 {
			r.app.display.ShowInfo("No commit suggestions found")
			return nil
		}

		if len(staged) > 0 {
			r.displayStagedChanges(staged)
		}

		r.displayCommitSuggestions(suggestions)
		message, regenrete, err := r.getCommitMessage(suggestions.Suggestions)
		if err != nil {
			return err
		}
		if regenrete {
			continue
		}

		if message == "" {
			r.app.display.ShowInfo("Commit cancelled")
			return nil
		}

		if err := r.app.commitAgent.Commit(ctx, message); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}

		r.app.display.ShowSuccess(fmt.Sprintf("Changes committed successfully with message: %s", message))
		return nil

	}
}

func (r *REPL) displayStagedChanges(changes []git.FileChange) {

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

func (r *REPL) displayCommitSuggestions(suggestions *agent.CommitResponse) {

	// Display summary section
	r.app.display.ShowSection(display.Section{
		Title:   "Change Summary",
		Icon:    "📝",
		Content: suggestions.Summary,
	})

	// Display suggestions section
	r.app.display.ShowSection(display.Section{
		Title: "Suggested Commit Messages",
		Icon:  "💡",
	})

	// Convert suggestions to list items
	var items []display.ListItem
	for _, suggestion := range suggestions.Suggestions {
		items = append(items, display.ListItem{
			Content:     suggestion.Message,
			Description: suggestion.Description,
		})
	}
	r.app.display.ShowNumberedList(items)
}

func (r *REPL) getCommitMessage(suggestions []agent.CommitSuggestion) (string, bool, error) {
	fmt.Print("\nSelect a message (1-3), 'r' to regenerate, 'c' to cancel, or 'm' for manual input: ")
	input, err := r.reader.ReadString('\n')
	if err != nil {
		return "", false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	switch input {
	case "c":
		return "", false, nil
	case "m":
		fmt.Print("Enter your commit message: ")
		message, err := r.reader.ReadString('\n')
		if err != nil {
			return "", false, fmt.Errorf("failed to read input: %w", err)
		}
		return strings.TrimSpace(message), false, nil

	case "r":
		return "", true, nil
	case "":
		return "", false, nil
	default:
		selection := 0
		if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
			return "", false, fmt.Errorf("invalid selection")
		}
		if selection < 1 || selection > len(suggestions) {
			return "", false, fmt.Errorf("invalid selection")
		}
		return suggestions[selection-1].Message, false, nil
	}
}

func getStatusSymbol(status string) string {
	switch status {
	case "modified":
		return "📝"
	case "added":
		return "➕"
	case "deleted":
		return "➖"
	case "renamed":
		return "📋"
	case "copied":
		return "📑"
	case "untracked":
		return "❓"
	default:
		return "•"
	}
}

func (r *REPL) handleChangeDirectory(input string) error {
	path := strings.TrimPrefix(input, "cd")
	path = strings.TrimSpace(path)

	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = strings.Replace(path, "~", homeDir, 1)
	}

	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	pwd, _ := os.Getwd()
	r.app.display.ShowInfo(fmt.Sprintf("Changed to: %s", pwd))
	r.app.agent.ResetChat()
	return nil
}
