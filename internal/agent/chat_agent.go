package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-coders/git_gpt/pkg/apierrors"
)

type ChatAgent struct {
	*BaseAgent
}

func NewChatAgent(config AgentConfig) (*ChatAgent, error) {
	base, err := NewBaseAgent(config)
	if err != nil {
		return nil, err
	}

	agent := &ChatAgent{
		BaseAgent: base,
	}

	if err := agent.ResetChat(); err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *ChatAgent) Chat(ctx context.Context, query string) error {
	if !a.git.IsGitRepository(ctx) {
		return apierrors.NewNotGitRepoError()
	}

	response, err := a.getCommandResponse(ctx, query)
	if err != nil {
		return err
	}

	return a.handleResponse(ctx, query, response)
}

func (a *ChatAgent) getCommandResponse(ctx context.Context, query string) (Response, error) {
	prompt, err := a.prompts.GetGenerateCommandsPrompt(query)
	if err != nil {
		return Response{}, fmt.Errorf("failed to generate prompt: %w", err)
	}

	a.display.StartSpinner("Analyzing query...")
	llmResponse, err := a.llm.Chat(ctx, prompt)
	a.display.StopSpinner()

	if err != nil {
		return Response{}, fmt.Errorf("LLM error: %w", err)
	}

	var response Response
	if err := json.Unmarshal([]byte(cleanJSONResponse(llmResponse)), &response); err != nil {
		return Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
}

func (a *ChatAgent) handleResponse(ctx context.Context, query string, response Response) error {
	switch response.Type {
	case "answer":
		a.display.ShowSuccess(response.Content)
		return nil

	case "execute":
		return a.handleExecuteResponse(ctx, query, response)

	default:
		return fmt.Errorf("unknown response type: %s", response.Type)
	}
}

func (a *ChatAgent) handleExecuteResponse(ctx context.Context, query string, response Response) error {
	switch response.CommandType {
	case CommandTypeQuery:
		return a.handleQueryCommands(ctx, query, response.Commands)

	case CommandTypeModify:
		return a.handleModificationCommands(ctx, response.Commands)

	default:
		return fmt.Errorf("unknown command type: %s", response.CommandType)
	}
}

func (a *ChatAgent) handleQueryCommands(ctx context.Context, query string, commands []Command) error {
	a.logger.Debug("handleQueryCommands,Executing commands: %v", commands)
	results, err := a.executeCommands(ctx, commands)
	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return a.summarizeResults(ctx, query, results)
}

func (a *ChatAgent) summarizeResults(ctx context.Context, query string, results []CommandResult) error {
	var builder strings.Builder
	for _, result := range results {
		builder.WriteString(fmt.Sprintf("Command: git %s\n", strings.Join(result.Command.Args, " ")))
		builder.WriteString(fmt.Sprintf("Output:\n%s\n\n", result.Output))
	}

	prompt, err := a.prompts.GetSummarizeResultsPrompt(query, builder.String())
	if err != nil {
		return fmt.Errorf("failed to generate summary prompt: %w", err)
	}

	a.display.StartSpinner("Analyzing results...")
	summary, err := a.llm.Chat(ctx, prompt)
	a.display.StopSpinner()

	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	a.display.ShowSuccess(summary)
	return nil
}

// handleModificationCommands handles commands that modify the repository
func (a *ChatAgent) handleModificationCommands(ctx context.Context, commands []Command) error {
	a.display.ShowWarning("The following commands will modify the repository:")

	// Show commands and their resolved forms
	for i, cmd := range commands {
		fmt.Println()
		cmdStr := fmt.Sprintf("git %s", strings.Join(cmd.Args, " "))
		a.display.ShowInfo(fmt.Sprintf("Command %d: %s", i+1, cmdStr))

		// Try to resolve any command substitutions for preview
		resolvedArgs, err := a.processShellCommandArgs(ctx, cmd.Args)
		if err == nil && !reflect.DeepEqual(resolvedArgs, cmd.Args) {
			a.display.ShowInfo(fmt.Sprintf("Will execute as: git %s",
				strings.Join(resolvedArgs, " ")))
		}

		if cmd.Purpose != "" {
			a.display.ShowInfo(fmt.Sprintf("Purpose: %s", cmd.Purpose))
		}
		if cmd.Impact != "" {
			a.display.ShowWarning(fmt.Sprintf("Impact: %s", cmd.Impact))
		}
	}

	// Get user confirmation
	confirmed, err := a.promptForConfirmation("\nDo you want to execute these commands? (y/n): ")
	if err != nil {
		return err
	}

	if !confirmed {
		a.display.ShowInfo("Operation cancelled")
		return nil
	}

	// Execute commands
	results, err := a.executeCommands(ctx, commands)
	if err != nil {
		return err
	}

	return a.handleCommandResults(ctx, results)
}
func (a *ChatAgent) ResetChat() error {
	systemPrompt, err := a.prompts.GetSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to get system prompt: %w", err)
	}

	a.llm.SetSystemMessage(systemPrompt)
	a.llm.ClearHistory()
	return nil
}
