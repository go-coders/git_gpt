// internal/agent/agent.go
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-coders/gitchat/internal/display"
	"github.com/go-coders/gitchat/internal/git"
	"github.com/go-coders/gitchat/internal/llm"
	"github.com/go-coders/gitchat/pkg/apierrors"
	"github.com/go-coders/gitchat/pkg/utils"
)

type Response struct {
	Type     string       `json:"type"`
	Content  string       `json:"content"`
	Commands []GitCommand `json:"commands"`
	Reason   string       `json:"reason"`
}

type GitCommand struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Purpose string   `json:"purpose"`
}

type ConversationContext struct {
	Messages       []llm.ConversationMessage
	CommandHistory []CommandContext
}

type CommandContext struct {
	Command string
	Output  string
	Time    time.Time
}

type Agent struct {
	git     git.Executor // Changed from *git.Executor to git.Executor interface
	llm     *llm.Client
	logger  utils.Logger
	context *ConversationContext
	prompts *Prompts
	display display.Manager
}

func New(llm *llm.Client, git git.Executor, logger utils.Logger, display display.Manager) *Agent {
	agent := &Agent{
		git:     git,
		llm:     llm,
		logger:  logger,
		prompts: &DefaultPrompts,
		display: display,
	}
	agent.ResetChat()
	return agent
}

func (a *Agent) Chat(ctx context.Context, query string) error {

	// chekc if git repo
	if !a.git.IsGitRepository(ctx) {
		return apierrors.NewNotGitRepoError()
	}

	msg := llm.ConversationMessage{
		Role:    llm.RoleUser,
		Origin:  query,
		Content: fmt.Sprintf(a.prompts.GenerateCommands, query),
	}

	if err := a.prepareMessages(msg); err != nil {
		return fmt.Errorf("failed to prepare messages: %w", err)
	}

	a.display.StartSpinner("Thinking...")
	response, err := a.llm.Complete(ctx, a.context.Messages)
	a.display.StopSpinner()

	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	a.logger.Debug("Response: %s", response)
	a.addToContext(llm.RoleAssist, response)

	response = cleanJSONResponse(response)
	var parsedResponse Response
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

type CommandResult struct {
	Command GitCommand
	Output  string
	Error   error
}

func (a *Agent) validateGitCommand(cmd GitCommand) error {
	if cmd.Command != "git" {
		return fmt.Errorf("invalid command type: %s", cmd.Command)
	}

	if len(cmd.Args) == 0 {
		return fmt.Errorf("empty git command arguments")
	}

	// Validate and sanitize format args if present
	for i, arg := range cmd.Args {
		if strings.Contains(arg, "--pretty=format:") || strings.Contains(arg, "--format=") {
			cmd.Args[i] = strings.ReplaceAll(arg, "--pretty=format:", "--pretty=format:%h - %an, %ar : %s")
			cmd.Args[i] = strings.ReplaceAll(arg, "--format=", "--format=%h - %an, %ar : %s")
		}
	}

	return nil
}

func (a *Agent) provideFinalAnswer(ctx context.Context, query string, results []CommandResult) error {
	var contextBuilder strings.Builder
	for _, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("Command: git %s\n", strings.Join(result.Command.Args, " ")))
		contextBuilder.WriteString(fmt.Sprintf("Output:\n%s\n\n", result.Output))
	}
	prompt := fmt.Sprintf(a.prompts.SummarizeResults, query, contextBuilder.String())
	msg := llm.ConversationMessage{
		Role:    llm.RoleUser,
		Content: prompt,
		Origin:  contextBuilder.String(),
	}

	if err := a.prepareMessages(msg); err != nil {
		return err
	}

	a.display.StartSpinner("Analyzing results...")
	answer, err := a.llm.Complete(ctx, a.context.Messages)
	a.display.StopSpinner()

	if err != nil {
		return fmt.Errorf("failed to generate final answer: %w", err)
	}

	a.display.ShowSuccess(answer)
	return nil
}

func (a *Agent) addToContext(role, content string) {
	a.context.Messages = append(a.context.Messages, llm.ConversationMessage{
		Role:    role,
		Content: content,
	})
}

func (a *Agent) ResetChat() {
	a.context = &ConversationContext{
		Messages: []llm.ConversationMessage{{
			Role:    "system",
			Content: a.prompts.System,
		}},
	}
}

// Rest of the code remains the same...

func (a *Agent) prepareMessages(msg llm.ConversationMessage) error {
	a.context.Messages = append(a.context.Messages, msg)

	systeMsg := llm.ConversationMessage{
		Role:    llm.RoleSystem,
		Content: a.prompts.System,
	}
	var newMessages []llm.ConversationMessage
	newMessages = append(newMessages, systeMsg)

	// for max  user msg
	var max = 0
	for i := 0; i < len(a.context.Messages); i++ {
		if a.context.Messages[i].Role == llm.RoleUser {
			max = i
		}
	}

	var contextMsg = []llm.ConversationMessage{}
	for i := 0; i < len(a.context.Messages); i++ {
		var msg = a.context.Messages[i]
		if msg.Role == llm.RoleUser && msg.Origin != "" && i != max {
			msg.Content = msg.Origin
		}
		contextMsg = append(contextMsg, msg)

	}

	for i := len(contextMsg) - 1; i >= 0; i-- {
		msg := contextMsg[i]
		if msg.Role == llm.RoleSystem {
			continue
		}
		tokens, err := a.llm.CountTokens(append(newMessages, msg))
		if err != nil {
			return err
		}

		if tokens > a.llm.MaxTokens() {
			break
		}
		newMessages = append([]llm.ConversationMessage{msg}, newMessages...)
	}

	newMessages = append([]llm.ConversationMessage{systeMsg}, newMessages[:len(newMessages)-1]...)
	a.context.Messages = newMessages
	if len(newMessages) == 1 {
		currentToken, _ := a.llm.CountTokens(append(newMessages, llm.ConversationMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}))
		return apierrors.NewTokenLimitError(currentToken, a.llm.MaxTokens())

	}

	return nil

}
