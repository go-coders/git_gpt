// internal/agent/prompts.go
package agent

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/go-coders/gitchat/internal/git"
)

// TemplateData holds all possible data that can be used in templates
type TemplateData struct {
	TimeContext    TimeContext
	Query          string
	Changes        []git.FileChange
	Diff           string
	CommandResults string
}

type TimeContext struct {
	CurrentTime   string
	Today         string
	Yesterday     string
	LastWeekStart string
	LastMonth     string
}

// Predefined templates
const (
	systemPromptTpl = `You are a Git expert assistant with deep understanding of version control systems.
You have extensive experience with git internals, workflows, and best practices.
You provide accurate, technically sound advice and commands.
Your responses are clear, direct and precise.

Current time context:
- Current time: {{.TimeContext.CurrentTime}}
- Today's date: {{.TimeContext.Today}}
- Yesterday: {{.TimeContext.Yesterday}}
- Last week start: {{.TimeContext.LastWeekStart}}
- Last month: {{.TimeContext.LastMonth}}

When analyzing queries:
1. Try to use existing command output first if available in the conversation
2. Only request new git commands if the information is not available
3. Be specific about what additional information you need and why
4. Use the current time context for relative time references
5. Always respond in the same language as the user's query`

	generateCommandsTpl = `Analyze the query and determine the appropriate git commands to execute.

Consider these command patterns:
1. For file list or general changes:
 - Use git log --name-status: lighter and shows changed files
 - Example: git log --name-status --since="3.days.ago" --until="1.day.ago"

2. For content changes:
 - Use git log -p for specific file or content changes
 - Example: git log -p --since="3.days.ago" path/to/file

3. For commit info only:
 - Use git log with format string 
 
Return a JSON response in one of these two formats:

If you can answer using existing information:
{
    "type": "answer",
    "content": "Your detailed answer based on the context"
}

If you need to execute commands:
{
    "type": "execute",
    "commands": [
        {
            "command": "git",
            "args": ["command", "args"],
            "purpose": "explain why this command is needed"
        }
    ],
    "reason": "Explain why these commands are needed"
}

Query: {{.Query}}`

	summarizeResultsTpl = `Answer this Git repository question based on the command results:

Question: {{.Query}}

Git command execution results:
{{.CommandResults}}

Instructions:
1. If any command output is empty, mention that no changes/data were found
2. Provide a concise answer that directly addresses the question
3. Answer in the same language as the question
4. Use plain text format, no formatting
5. If the output suggests an error, explain it simply
6. Keep technical details only if directly relevant`

	commitPromptTpl = `Analyze these git changes and generate commit message suggestions.
Return a JSON response in this exact format:
{
    "summary": "A brief summary of the changes in markdown format",
    "suggestions": [
        {
            "message": "type(scope): subject"
        }
    ]
}

Changes:
{{range .Changes}}- {{.Status}}: {{.Path}} ({{.Additions}}+/{{.Deletions}}-)
{{end}}

Detailed diff:
{{.Diff}}

Guidelines for commit messages:
1. Use conventional commits format: type(scope): description
2. Available types: feat, fix, docs, style, refactor, test, chore
4. Focus on what changes accomplish, not how
5. No period at the end
6. Use imperative mood ("add" not "added")
7. Generate exactly 3 different suggestions
8. Each suggestion should focus on a different aspect

Guidelines for summary:
1. Brief but comprehensive summary of changes
2. Focus on the overall impact
3. Keep it under 3-4 sentences
4. Include key changes and their purposes
5. Use technical but clear language`
)

// PromptManager handles template rendering for different prompts
type PromptManager struct {
	systemPrompt     *template.Template
	generateCommands *template.Template
	summarizeResults *template.Template
	commitPrompt     *template.Template
}

func NewPromptManager() (*PromptManager, error) {
	pm := &PromptManager{}

	var err error
	if pm.systemPrompt, err = template.New("system").Parse(systemPromptTpl); err != nil {
		return nil, fmt.Errorf("failed to parse system prompt template: %w", err)
	}

	if pm.generateCommands, err = template.New("commands").Parse(generateCommandsTpl); err != nil {
		return nil, fmt.Errorf("failed to parse commands template: %w", err)
	}

	if pm.summarizeResults, err = template.New("summarize").Parse(summarizeResultsTpl); err != nil {
		return nil, fmt.Errorf("failed to parse summarize template: %w", err)
	}

	if pm.commitPrompt, err = template.New("commit").Parse(commitPromptTpl); err != nil {
		return nil, fmt.Errorf("failed to parse commit template: %w", err)
	}

	return pm, nil
}

func (pm *PromptManager) GetSystemPrompt() (string, error) {
	data := TemplateData{
		TimeContext: getTimeContext(),
	}
	return pm.renderTemplate(pm.systemPrompt, data)
}

func (pm *PromptManager) GetGenerateCommandsPrompt(query string) (string, error) {
	data := TemplateData{
		Query: query,
	}
	return pm.renderTemplate(pm.generateCommands, data)
}

func (pm *PromptManager) GetSummarizeResultsPrompt(query, results string) (string, error) {
	data := TemplateData{
		Query:          query,
		CommandResults: results,
	}
	return pm.renderTemplate(pm.summarizeResults, data)
}

func (pm *PromptManager) GetCommitPrompt(changes []git.FileChange, diff string) (string, error) {
	data := TemplateData{
		Changes: changes,
		Diff:    diff,
	}
	return pm.renderTemplate(pm.commitPrompt, data)
}

func (pm *PromptManager) renderTemplate(tmpl *template.Template, data TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}
	return buf.String(), nil
}

func getTimeContext() TimeContext {
	now := time.Now()
	return TimeContext{
		CurrentTime:   now.Format("15:04:05"),
		Today:         now.Format("2006-01-02"),
		Yesterday:     now.AddDate(0, 0, -1).Format("2006-01-02"),
		LastWeekStart: now.AddDate(0, 0, -7).Format("2006-01-02"),
		LastMonth:     now.AddDate(0, -1, 0).Format("2006-01-02"),
	}
}
