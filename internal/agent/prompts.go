package agent

import (
	"fmt"
	"time"
)

type Prompts struct {
	System           string
	GenerateCommands string
	SummarizeResults string
}

var DefaultPrompts = Prompts{}

func getTimeContext() string {
	now := time.Now()
	return fmt.Sprintf(`Current time context:
- Current time: %s
- Today's date: %s
- Yesterday: %s
- Last week start: %s
- Last month: %s`,
		now.Format("15:04:05"),
		now.Format("2006-01-02"),
		now.AddDate(0, 0, -1).Format("2006-01-02"),
		now.AddDate(0, 0, -7).Format("2006-01-02"),
		now.AddDate(0, -1, 0).Format("2006-01-02"))
}

func init() {
	timeContext := getTimeContext()

	DefaultPrompts.System = fmt.Sprintf(`You are a Git expert assistant with deep understanding of version control systems.
You have extensive experience with git internals, workflows, and best practices.
You provide accurate, technically sound advice and commands.
Your responses are clear, direct and precise.

%s

When analyzing queries:
1. Try to use existing command output first if available in the conversation
2. Only request new git commands if the information is not available
3. Be specific about what additional information you need and why
4. Use the current time context for relative time references
5. Always respond in the same language as the user's query`, timeContext)

	DefaultPrompts.GenerateCommands = `Analyze the query and determine the appropriate git commands to execute.

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

Query: %s`

	DefaultPrompts.SummarizeResults = `Answer this Git repository question based on the command results:

Question: %s

Git command execution results:
%s

Instructions:
1. If any command output is empty, mention that no changes/data were found
2. Provide a concise answer that directly addresses the question
3. Answer in the same language as the question
4. Use plain text format, no formatting
5. If the output suggests an error, explain it simply
6. Keep technical details only if directly relevant`
}
