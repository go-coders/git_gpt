// internal/agent/agent.go
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-coders/gitchat/pkg/apierrors"
)

type Agent struct {
	git     GitExecutor
	llm     LLMClient
	logger  Logger
	display DisplayManager
	prompts *PromptManager
}

func New(llm LLMClient, git GitExecutor, logger Logger, display DisplayManager) (*Agent, error) {

	prompts, err := NewPromptManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt manager: %w", err)
	}
	agent := &Agent{
		git:     git,
		llm:     llm,
		logger:  logger,
		display: display,
		prompts: prompts,
	}
	agent.ResetChat()
	return agent, nil
}

func (a *Agent) Chat(ctx context.Context, query string) error {
	if !a.git.IsGitRepository(ctx) {
		return apierrors.NewNotGitRepoError()
	}

	prompt, err := a.prompts.GetGenerateCommandsPrompt(query)

	if err != nil {
		return fmt.Errorf("failed to generate prompt: %w", err)
	}

	a.display.StartSpinner("Thinking...")
	response, err := a.llm.Chat(ctx, prompt)
	a.display.StopSpinner()

	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	a.logger.Debug("Response: %s", response)

	var parsedResponse Response
	response = cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(response), &parsedResponse); err != nil {
		return fmt.Errorf("invalid response format: %w", err)
	}

	return a.handleResponse(ctx, query, parsedResponse)
}
func (a *Agent) handleResponse(ctx context.Context, query string, response Response) error {
	switch response.Type {
	case "answer":
		a.display.ShowSuccess(response.Content)
		return nil
	case "execute":
		results, err := a.executeCommands(ctx, response.Commands)
		if err != nil {
			return fmt.Errorf("failed to execute commands: %w", err)
		}
		return a.provideFinalAnswer(ctx, query, results)
	default:
		return fmt.Errorf("unknown response type: %s", response.Type)
	}
}

func (a *Agent) provideFinalAnswer(ctx context.Context, query string, results []CommandResult) error {
	var contextBuilder strings.Builder
	for _, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("Command: git %s\n", strings.Join(result.Command.Args, " ")))
		contextBuilder.WriteString(fmt.Sprintf("Output:\n%s\n\n", result.Output))
	}

	prompt, err := a.prompts.GetSummarizeResultsPrompt(query, contextBuilder.String())
	if err != nil {
		return fmt.Errorf("failed to generate summary prompt: %w", err)
	}

	a.display.StartSpinner("Analyzing results...")
	answer, err := a.llm.Chat(ctx, prompt)
	a.display.StopSpinner()
	if err != nil {
		return fmt.Errorf("failed to generate final answer: %w", err)
	}

	a.display.ShowSuccess(answer)
	return nil
}

func (a *Agent) ResetChat() error {
	systemPrompt, err := a.prompts.GetSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to get system prompt: %w", err)
	}

	a.llm.SetSystemMessage(systemPrompt)
	a.llm.ClearHistory()
	return nil
}

func (a *Agent) executeCommands(ctx context.Context, commands []GitCommand) ([]CommandResult, error) {
	var results []CommandResult
	for _, cmd := range commands {
		if err := a.validateGitCommand(cmd); err != nil {
			return nil, fmt.Errorf("invalid git command: %w", err)
		}

		// Show the command being executed
		cmdStr := fmt.Sprintf("git %s", strings.Join(cmd.Args, " "))
		a.display.ShowCommand(cmdStr)

		// Execute command
		output, err := a.git.Execute(ctx, cmd.Args...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute git command '%s': %w", cmdStr, err)
		}

		results = append(results, CommandResult{
			Command: cmd,
			Output:  output,
			Error:   err,
		})
	}
	return results, nil
}

func (a *Agent) validateGitCommand(cmd GitCommand) error {
	if cmd.Command != "git" {
		return fmt.Errorf("invalid command type: %s", cmd.Command)
	}

	if len(cmd.Args) == 0 {
		return fmt.Errorf("empty git command arguments")
	}

	// Validate and sanitize format args if present
	// for i, arg := range cmd.Args {
	// 	if strings.Contains(arg, "--pretty=format:") || strings.Contains(arg, "--format=") {
	// 		cmd.Args[i] = strings.ReplaceAll(arg, "--pretty=format:", "--pretty=format:%h - %an, %ar : %s")
	// 		cmd.Args[i] = strings.ReplaceAll(arg, "--format=", "--format=%h - %an, %ar : %s")
	// 	}
	// }

	return nil
}
