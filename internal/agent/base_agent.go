package agent

import (
	"bufio"
	"context"
	"fmt"
	"reflect"
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

// executeCommands executes a sequence of commands
func (a *BaseAgent) executeCommands(ctx context.Context, commands []Command) ([]CommandResult, error) {
	results := make([]CommandResult, 0, len(commands))

	for _, cmd := range commands {
		a.display.ShowCommand(fmt.Sprintf("git %s", strings.Join(cmd.Args, " ")))
		result, err := a.executeSingleCommand(ctx, cmd)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// processShellCommandArgs handles shell command substitution patterns flexibly
func (a *BaseAgent) processShellCommandArgs(ctx context.Context, args []string) ([]string, error) {
	processedArgs := make([]string, 0, len(args))

	for _, arg := range args {
		// If argument contains shell substitution
		if strings.Contains(arg, "$(") && strings.Contains(arg, ")") {
			// Process all command substitutions in this argument
			processedArg, err := a.processSubstitutions(ctx, arg)
			if err != nil {
				return nil, err
			}
			processedArgs = append(processedArgs, processedArg)
		} else {
			processedArgs = append(processedArgs, arg)
		}
	}

	return processedArgs, nil
}

// extractSubstitutions finds all command substitutions in a string
func extractSubstitutions(s string) []string {
	var substitutions []string
	var depth, start int
	inQuotes := false
	quoteChar := rune(0)

	for i, r := range s {
		switch r {
		case '\'', '"':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
		case '$':
			if !inQuotes && i+1 < len(s) && s[i+1] == '(' {
				if depth == 0 {
					start = i + 2 // skip $(
				}
				depth++
			}
		case ')':
			if !inQuotes && depth > 0 {
				depth--
				if depth == 0 && start < i {
					substitutions = append(substitutions, strings.TrimSpace(s[start:i]))
				}
			}
		}
	}

	return substitutions
}

// parseCommandString parses a command string into parts, handling quotes correctly
func parseCommandString(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range cmd {
		switch {
		case r == '"' || r == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// handleModificationCommands handles commands that modify the repository

// processSubstitutions handles all command substitutions in a single argument
func (a *BaseAgent) processSubstitutions(ctx context.Context, arg string) (string, error) {
	result := arg

	// Find all $(git ...) patterns
	substitutions := extractSubstitutions(arg)

	// Process each substitution
	for _, sub := range substitutions {
		// Only process git commands
		if !strings.HasPrefix(strings.TrimSpace(sub), "git ") {
			continue
		}

		// Extract the command parts (remove the initial 'git' command)
		cmdParts := parseCommandString(strings.TrimPrefix(sub, "git"))
		if len(cmdParts) == 0 {
			continue
		}

		// Special handling for git describe command
		if cmdParts[0] == "describe" && contains(cmdParts, "--tags") {
			// Correct the command to use proper git arguments
			describeArgs := []string{"describe", "--tags"}
			if contains(cmdParts, "--abbrev=0") {
				describeArgs = append(describeArgs, "--abbrev=0")
			}

			// Execute the corrected describe command
			output, err := a.git.Execute(ctx, describeArgs...)
			if err != nil {
				return "", fmt.Errorf("failed to get tag description: %w", err)
			}

			// Replace the substitution with the command output
			output = strings.TrimSpace(output)
			result = strings.ReplaceAll(result, fmt.Sprintf("$(%s)", sub), output)
			continue
		}

		// Log the substitution being processed
		a.logger.Debug("Processing substitution: %v", cmdParts)

		// Execute the nested command
		output, err := a.git.Execute(ctx, cmdParts...)
		if err != nil {
			return "", fmt.Errorf("failed to execute substitution '%s': %w", sub, err)
		}

		// Replace the substitution with the command output
		output = strings.TrimSpace(output)
		result = strings.ReplaceAll(result, fmt.Sprintf("$(%s)", sub), output)
	}

	return result, nil
}

// Helper function to check if slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// For better error handling, let's also modify executeSingleCommand
func (a *BaseAgent) executeSingleCommand(ctx context.Context, cmd Command) (CommandResult, error) {
	if len(cmd.Args) == 0 {
		return CommandResult{}, fmt.Errorf("empty command arguments")
	}

	// Process any shell command substitutions
	resolvedArgs, err := a.processShellCommandArgs(ctx, cmd.Args)
	if err != nil {
		if strings.Contains(err.Error(), "failed to get tag description") {
			a.display.ShowWarning("No tags found or error accessing tags")
			return CommandResult{}, fmt.Errorf("no tags found or error accessing tags")
		}
		return CommandResult{}, fmt.Errorf("failed to process command: %w", err)
	}

	// Log the resolved command
	if !reflect.DeepEqual(resolvedArgs, cmd.Args) {
		a.logger.Debug("Original command: %v", cmd.Args)
		a.logger.Debug("Resolved command: %v", resolvedArgs)
	}

	// Execute the command with resolved arguments
	output, err := a.git.Execute(ctx, resolvedArgs...)

	return CommandResult{
		Command: Command{
			Type:    cmd.Type,
			Args:    resolvedArgs,
			Purpose: cmd.Purpose,
			Impact:  cmd.Impact,
		},
		Output: output,
		Error:  err,
	}, err
}
