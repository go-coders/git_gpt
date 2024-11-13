package app

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/go-coders/git_gpt/internal/agent"
	"github.com/go-coders/git_gpt/internal/config"
	"github.com/go-coders/git_gpt/internal/display"
	"github.com/go-coders/git_gpt/internal/git"
	"github.com/go-coders/git_gpt/internal/llm"
	"github.com/go-coders/git_gpt/pkg/utils"
)

// Application represents the main application instance with its dependencies
type Application struct {
	config      *config.Config
	display     *display.DisplayImpl
	logger      *utils.LoggerImpl
	version     string
	gitClient   *git.GitExecutor
	chatAgent   *agent.ChatAgent
	commitAgent *agent.CommitAgent
	repl        *REPL
	mu          sync.RWMutex
}

// Options contains initialization parameters
type Options struct {
	Config  *config.Config
	Logger  *utils.LoggerImpl
	Version string
}

// New creates a new Application instance
func New(opts Options) (*Application, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	app := &Application{
		config:    opts.Config,
		logger:    opts.Logger,
		version:   opts.Version,
		display:   display.NewManager(opts.Version),
		gitClient: git.NewExecutor(),
	}

	if err := app.initialize(); err != nil {
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	return app, nil
}

func (a *Application) Run(ctx context.Context) error {
	a.display.ShowWelcome()
	return a.repl.Start(ctx)
}

func (a *Application) Reload() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	newConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	a.config = newConfig
	return a.initialize()
}

func (a *Application) HandleErr(err error) {
	if err == nil {
		return
	}

	a.logger.Error("error occurred: %v", err)
	a.display.ShowError(err.Error())
}

func (a *Application) GetConfig() *config.Config {
	return a.config
}

// Private initialization methods

func validateOptions(opts Options) error {
	if opts.Config == nil {
		return fmt.Errorf("config is required")
	}
	if opts.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	return nil
}

func (a *Application) initialize() error {
	if err := a.initializeLLMServices(); err != nil {
		return err
	}

	a.repl = NewREPL(a)
	return nil
}

func (a *Application) initializeLLMServices() error {
	if a.config.LLM.APIKey == "" {
		if err := a.runConfigWizard(); err != nil {
			return err
		}
	}

	chatLLM, commitLLM, err := a.createLLMClients()
	if err != nil {
		return err
	}

	if err := a.createAgents(chatLLM, commitLLM); err != nil {
		return err
	}

	return nil
}

func (a *Application) createLLMClients() (chatLLM, commitLLM *llm.Client, err error) {
	// Create chat LLM client with history enabled
	chatLLM, err = llm.NewClient(llm.Config{
		APIKey:        a.config.LLM.APIKey,
		BaseURL:       a.config.LLM.BaseURL,
		Model:         a.config.LLM.Model,
		MaxTokens:     a.config.LLM.MaxTokens,
		Temperature:   a.config.LLM.ChatTemperture,
		EnableHistory: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize chat LLM client: %w", err)
	}

	// Create commit LLM client without history
	commitLLM, err = llm.NewClient(llm.Config{
		APIKey:        a.config.LLM.APIKey,
		BaseURL:       a.config.LLM.BaseURL,
		Model:         a.config.LLM.Model,
		Temperature:   a.config.LLM.CommitTemperture,
		EnableHistory: false,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize commit LLM client: %w", err)
	}

	return chatLLM, commitLLM, nil
}

func (a *Application) createAgents(chatLLM, commitLLM *llm.Client) error {
	// Create base config for agents
	baseConfig := agent.AgentConfig{
		Git:     a.gitClient,
		Display: a.display,
		Logger:  a.logger,
		Reader:  os.Stdin,
	}

	// Create chat agent
	chatConfig := baseConfig
	chatConfig.LLM = chatLLM
	chat, err := agent.NewChatAgent(chatConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize chat agent: %w", err)
	}

	// Create commit agent
	commitConfig := baseConfig
	commitConfig.LLM = commitLLM
	commit, err := agent.NewCommitAgent(commitConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize commit agent: %w", err)
	}

	a.chatAgent = chat
	a.commitAgent = commit

	return nil
}

func (a *Application) runConfigWizard() error {
	wizard := NewConfigWizard(a.config)
	if err := wizard.Run(); err != nil {
		return fmt.Errorf("configuration wizard failed: %w", err)
	}
	return nil
}
