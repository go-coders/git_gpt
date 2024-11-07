package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-coders/gitchat/internal/agent"
	"github.com/go-coders/gitchat/internal/config"
	"github.com/go-coders/gitchat/internal/display"
	"github.com/go-coders/gitchat/internal/git"
	"github.com/go-coders/gitchat/internal/llm"
	"github.com/go-coders/gitchat/pkg/apierrors"
	"github.com/go-coders/gitchat/pkg/utils"
)

type Application struct {
	config      *config.Config
	display     display.Manager
	logger      utils.Logger
	repl        *REPL
	version     string
	gitClient   git.Executor
	agent       *agent.Agent
	commitAgent *agent.CommitAgent
}

func New(cfg *config.Config, logger utils.Logger, version string) (*Application, error) {
	app := &Application{
		logger:    logger,
		config:    cfg,
		version:   version,
		display:   display.NewManager(version),
		gitClient: git.NewExecutor(logger),
	}
	app.repl = NewREPL(app)
	return app, nil
}

func (app *Application) loadLlm() error {
	if app.config.LLM.APIKey == "" {
		wizard := NewConfigWizard(app.config, app.logger)
		if err := wizard.Run(); err != nil {
			return fmt.Errorf("failed to run configuration wizard: %w", err)
		}
	}
	llmClient, err := llm.NewClient(llm.ClientConfig{
		APIKey:    app.config.LLM.APIKey,
		Model:     app.config.LLM.Model,
		BaseURL:   app.config.LLM.BaseURL,
		MaxTokens: app.config.LLM.MaxTokens,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize LLM client: %w", err)
	}
	app.agent = agent.New(llmClient, app.gitClient, app.logger, app.display)
	app.commitAgent = agent.NewCommitAgent(app.gitClient, llmClient, app.display, app.logger)

	return nil
}

func (app *Application) Run(ctx context.Context) error {
	err := app.loadLlm()
	if err != nil {
		return fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	app.display.ShowWelcome()
	if err := app.repl.Start(ctx); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}

// Reload reinitializes the application with updated configuration
func (app *Application) Reload() error {
	newConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}
	app.config = newConfig
	return app.loadLlm()
}

// GetConfig returns the current configuration
func (app *Application) GetConfig() *config.Config {
	return app.config
}

func (h *Application) HandleErr(err error) {
	if err == nil {
		return
	}
	h.logger.Error("error occurred: %v", err)
	var appErr *apierrors.AppError
	if errors.As(err, &appErr) {
		h.display.ShowError(appErr.Message)
		return
	}
	// show generic error
	h.display.ShowError(err.Error())

}
