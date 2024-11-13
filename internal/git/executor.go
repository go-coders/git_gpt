package git

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-coders/git_gpt/internal/common"
)

// GitExecutor implements the Executor interface
type GitExecutor struct {
}

// NewExecutor creates a new GitExecutor instance
func NewExecutor() *GitExecutor {
	return &GitExecutor{}
}

func (e *GitExecutor) Execute(ctx context.Context, args ...string) (string, error) {

	cmd := exec.CommandContext(ctx, "git", args...)

	env := append(cmd.Env, "GIT_PAGER=cat", "PAGER=cat", "GIT_TERMINAL_PROMPT=0")
	cmd.Env = env
	cmd.Stdin = os.Stdin

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w: %s", err, string(output))
	}

	result := cleanOutput(string(output))
	return result, nil
}

func cleanOutput(output string) string {
	// Remove common terminal control sequences
	output = strings.ReplaceAll(output, "\r", "")

	// Remove (END) marker that sometimes appears in paged output
	output = strings.TrimSuffix(output, "(END)")
	output = strings.TrimSuffix(output, "(END)\n")

	// Normalize newlines
	output = strings.ReplaceAll(output, "\r\n", "\n")

	// Trim trailing spaces from each line
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	output = strings.Join(lines, "\n")

	// Trim trailing newlines from the entire output
	output = strings.TrimRight(output, "\n")

	return output
}

func (e *GitExecutor) IsGitRepository(ctx context.Context) bool {
	_, err := e.Execute(ctx, "rev-parse", "--git-dir")
	return err == nil
}

func (e *GitExecutor) GetCurrentBranch(ctx context.Context) (string, error) {
	return e.Execute(ctx, "branch", "--show-current")
}

func (e *GitExecutor) StageAll(ctx context.Context) error {
	_, err := e.Execute(ctx, "add", "-A")
	if err != nil {
		return fmt.Errorf("failed to stage all changes: %w", err)
	}
	return nil
}

func (e *GitExecutor) StageFiles(ctx context.Context, files []string) error {
	args := append([]string{"add"}, files...)
	_, err := e.Execute(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}
	return nil
}

func (e *GitExecutor) Commit(ctx context.Context, message string) error {
	_, err := e.Execute(ctx, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func (e *GitExecutor) GetDiff(ctx context.Context, staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}

	output, err := e.Execute(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return output, nil
}

func (e *GitExecutor) GetStatus(ctx context.Context) (staged, unstaged []common.FileChange, err error) {
	statusResult, err := e.Execute(ctx, "status", "--porcelain")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get status: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(statusResult))
	// Regular expression to match status lines, including untracked files
	re := regexp.MustCompile(`^(?P<status>[ MADRCU?!]{2})\s+(?P<path>.+)$`)

	for scanner.Scan() {
		rawLine := scanner.Text()

		if len(rawLine) < 3 {
			continue
		}

		matches := re.FindStringSubmatch(rawLine)
		if matches == nil {
			continue
		}

		statusCode := matches[1]
		path := matches[2]

		indexStatus := statusCode[0]
		workingStatus := statusCode[1]

		// Determine if the change is staged based on indexStatus.
		if indexStatus != ' ' && indexStatus != '?' {
			change := common.FileChange{
				Path:   path,
				Status: getReadableStatus(indexStatus),
			}
			staged = append(staged, change)
		}

		// Determine if the change is unstaged based on workingStatus.
		if workingStatus != ' ' && workingStatus != '?' {
			change := common.FileChange{
				Path:   path,
				Status: getReadableStatus(workingStatus),
			}
			unstaged = append(unstaged, change)
		}

		// Handle untracked files
		if indexStatus == '?' && workingStatus == '?' {
			change := common.FileChange{
				Path:   path,
				Status: "untracked",
			}

			unstaged = append(unstaged, change)
		}
	}

	// Get stats for staged files if any.
	if len(staged) > 0 {
		statsResult, err := e.Execute(ctx, "diff", "--cached", "--numstat")
		if err == nil {
			stats := make(map[string]struct{ add, del int })
			scanner := bufio.NewScanner(strings.NewReader(statsResult))
			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					add, _ := strconv.Atoi(parts[0])
					del, _ := strconv.Atoi(parts[1])
					path := strings.Join(parts[2:], " ")
					stats[path] = struct{ add, del int }{add, del}
				}
			}

			// Update stats for staged files.
			for i := range staged {
				if stat, ok := stats[staged[i].Path]; ok {
					staged[i].Additions = stat.add
					staged[i].Deletions = stat.del
				}
			}
		}
	}

	return staged, unstaged, nil
}

func getReadableStatus(code byte) string {
	switch code {
	case 'M':
		return "modified"
	case 'A':
		return "added"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	case 'C':
		return "copied"
	case '?':
		return "untracked"
	case '!':
		return "ignored"
	case 'U':
		return "unmerged"
	default:
		return "unknown"
	}
}
