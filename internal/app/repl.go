// app/repl.go
package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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
		return r.handleConfig()
	case input == "commit":
		return r.app.commitAgent.HandleCommit(ctx)
	case strings.HasPrefix(input, "cd"):
		return r.handleChangeDirectory(input)
	case strings.TrimSpace(input) == "":
		return nil
	default:
		return r.app.agent.Chat(ctx, input)
	}
}

func (r *REPL) handleConfig() error {
	wizard := NewConfigWizard(r.app.config)
	if err := wizard.Run(); err != nil {
		return err
	}
	if err := r.app.Reload(); err != nil {
		return fmt.Errorf("failed to reload application: %w", err)
	}
	r.app.display.ShowSuccess("Configuration updated and reloaded successfully")
	return nil
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
