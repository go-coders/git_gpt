package agent

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	git     GitExecutor
	llm     LLMClient
	display DisplayManager
	logger  Logger
	reader  *bufio.Reader
	prompts *PromptManager
}

func NewBaseAgent(config AgentConfig) (*BaseAgent, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	prompts, err := NewPromptManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt manager: %w", err)
	}

	return &BaseAgent{
		git:     config.Git,
		llm:     config.LLM,
		display: config.Display,
		logger:  config.Logger,
		reader:  bufio.NewReader(config.Reader),
		prompts: prompts,
	}, nil
}

func validateConfig(config AgentConfig) error {
	if config.Git == nil {
		return fmt.Errorf("git executor is required")
	}
	if config.LLM == nil {
		return fmt.Errorf("LLM client is required")
	}
	if config.Display == nil {
		return fmt.Errorf("display manager is required")
	}
	if config.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	if config.Reader == nil {
		return fmt.Errorf("reader is required")
	}
	return nil
}

// Common functionality

func (a *BaseAgent) executeCommands(ctx context.Context, commands []Command) ([]CommandResult, error) {
	var results []CommandResult
	for _, cmd := range commands {
		output, err := a.git.Execute(ctx, cmd.Args...)
		results = append(results, CommandResult{
			Command: cmd,
			Output:  output,
			Error:   err,
		})

		if err != nil {
			return results, err
		}
	}
	return results, nil
}

func (a *BaseAgent) promptForConfirmation(prompt string) (bool, error) {
	fmt.Print(prompt)
	input, err := a.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(input) == "y", nil
}

func (a *BaseAgent) handleCommandResults(ctx context.Context, results []CommandResult) error {
	for _, result := range results {
		if result.Error != nil {
			a.display.ShowError(fmt.Sprintf("Command failed: %s", result.Error))
			return result.Error
		}
		cmdStr := fmt.Sprintf("git %s", strings.Join(result.Command.Args, " "))
		a.display.ShowSuccess(fmt.Sprintf("Executed: %s", cmdStr))
		if result.Output != "" {
			fmt.Println(result.Output)
		}
	}
	return nil
}

func cleanJSONResponse(response string) string {
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	return strings.TrimSpace(response)
}
