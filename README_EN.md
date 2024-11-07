# GitChat

English | [简体中文](README.md)

GitChat is a command-line tool for interacting with Git using natural language, designed to simplify Git operations and improve productivity. It leverages AI technology to understand natural language instructions, helping developers manage code changes and version history more efficiently.

<div align="center">

[![Release](https://img.shields.io/github/v/release/go-coders/gitchat)](https://github.com/go-coders/gitchat/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-coders/gitchat)](https://goreportcard.com/report/github.com/go-coders/gitchat)
[![License](https://img.shields.io/github/license/go-coders/gitchat)](LICENSE)

</div>

## 📦 Installation

### Using Go Install (requires Go 1.20+)

```bash
go install github.com/go-coders/gitchat/cmd/gitchat@latest
```

### Download Pre-compiled Binary

Go to the [Releases](https://github.com/go-coders/gitchat/releases) page, download the executable file suitable for your operating system, and add it to the system's PATH.

## 🚀 Quick Start

1. After installation, run in your terminal:

   ```bash
   gitchat
   ```

2. On first run, the configuration wizard will start. You'll need to provide:

   - OpenAI API key
   - Model selection (default: gpt-4o-mini)
   - API base URL (default: https://api.openai.com/v1)
   - Maximum tokens (default: 4000)

3. Once configured, you'll see the GitChat welcome screen!

```bash
🤖 Welcome to GitChat!
------------------------

  Natural Language  - Use natural language to interact with Git
                        使用自然语言与Git交互
  commit            - Generate commit message and commit changes
                        生成提交消息并提交更改
  config            - Run configuration wizard
                        运行配置向导
  cd <path>         - Change working directory
                        更改工作目录
  exit              - Quit the application
                        退出应用程序

```

## 💡 Usage Examples

### Natural Language Git Interaction

Use natural language to get repository insights:

```bash
> what files were modified in the last week

🔄 Executing: git log --name-status --since=2024-11-01
✅ Files modified in the last week:
- `README.md`
- `README_EN.md`
- `.goreleaser.yml`
- `cmd/main.go` (renamed to `cmd/gitchat/main.go`)

```

```bash
> write a 100-word daily report based on the last commit

🔄 Executing: git log -p -1
✅ Today's work focused on enhancing Git repository validation functionality. I added new code in chat_agent.go to verify whether the current directory is a Git repository before executing chat functionality. If not, it returns a custom NotGitRepoError. Additionally, I cleaned up the response handling to ensure proper formatting. These improvements enhance system robustness by preventing unnecessary operations in non-Git repository environments.

```

### Smart Commit Message Generation

When you want to commit your changes:

```bash
> commit
```

GitChat will analyze your changes and suggest appropriate commit messages:

```bash
📄 Staged files:
------------------------

📝 internal/agent/commit_agent.go (16+/18-)

📝 Change summary
------------------------
Enhanced the PrepareCommit function by adding checks for valid Git repositories and refactoring response handling. Introduced a new error type for non-Git repositories and modified the return type to include a structured CommitResponse. Improved logging and error handling in the generateSuggestions function to ensure clearer and more reliable suggestion generation.

💡 Suggested commit messages
------------------------
1) feat(agent): Add checks for valid Git repositories
2) refactor(agent): Update response handling in PrepareCommit
3) fix(agent): Improve error handling in suggestion generation

Please select a message (1-3), enter 'r' to regenerate, enter 'c' to cancel, or enter 'm' to manually input: 1
✅ Changes successfully committed, commit message: feat(agent): Add checks for valid Git repositories

```

## 📬 Contact & Support

- Report issues or suggest features on the [Issues](https://github.com/go-coders/gitchat/issues) page
- If you find it useful, please give us a Star!

---

Built with ❤️ by [Go Coders](https://github.com/go-coders)
