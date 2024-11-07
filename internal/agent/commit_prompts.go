// internal/agent/types.go
package agent

import (
	"fmt"
	"strings"

	"github.com/go-coders/gitchat/internal/git"
)

// CommitPrompt handles generation of prompts for commit message suggestions
type CommitPrompt struct {
	template string
}

func NewCommitPrompt() *CommitPrompt {
	return &CommitPrompt{
		template: `Analyze these git changes and generate commit message suggestions.
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
%s

Detailed diff:
%s

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
5. Use technical but clear language

Example response format:
{
    "summary": "Adds error handling for empty input in REPL. Improves user experience by gracefully handling blank inputs and providing clear feedback.",
    "suggestions": [
        {
            "message": "fix(repl): handle empty user input"
        }
    ]
}`,
	}
}

func (p *CommitPrompt) GeneratePrompt(changes []git.FileChange, diff string) string {
	return fmt.Sprintf(p.template,
		p.formatChanges(changes),
		diff,
	)
}

func (p *CommitPrompt) formatChanges(changes []git.FileChange) string {
	var result strings.Builder
	for _, change := range changes {
		result.WriteString(fmt.Sprintf("- %s: %s (%d+/%d-)\n",
			change.Status,
			change.Path,
			change.Additions,
			change.Deletions))
	}
	return result.String()
}
