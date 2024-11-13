# GitGPT

[简体中文](README.md) | English

GitGPT is an innovative command-line tool that seamlessly integrates GPT large language models with Git, enabling natural language interactions with Git. Without memorizing complex Git commands, you can perform Git operations through everyday conversations, such as "help me create a branch for login functionality" or "show me code changes from last week". It not only understands your intentions but also provides clear explanations and confirmations before executing critical operations, making Git operations more intelligent, secure, and efficient.

<div align="center">

[![Release](https://img.shields.io/github/v/release/go-coders/git_gpt)](https://github.com/go-coders/git_gpt/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-coders/git_gpt)](https://goreportcard.com/report/github.com/go-coders/git_gpt)
[![Tests](https://github.com/go-coders/git_gpt/actions/workflows/test.yml/badge.svg)](https://github.com/go-coders/git_gpt/actions/workflows/test.yml)
[![Coverage Status](https://codecov.io/gh/go-coders/git_gpt/branch/main/graph/badge.svg)](https://codecov.io/gh/go-coders/git_gpt)
[![License](https://img.shields.io/github/license/go-coders/git_gpt)](LICENSE)

</div>

## 📦 Installation

### Using Go Install (Requires Go 1.20+)

```bash
go install github.com/go-coders/git_gpt/cmd/ggpt@latest
```

### Download Pre-compiled Binaries

Visit the [Releases](https://github.com/go-coders/git_gpt/releases) page, download the executable file for your operating system, and add it to your system's PATH.

## 🚀 Quick Start

1. After installation, run in terminal:

   ```bash
   ggpt
   ```

2. On first run, a configuration wizard will start. You'll need to provide:

   - OpenAI API key
   - Model selection (default: gpt-4o)
   - API base URL (default: https://api.openai.com/v1)
   - Maximum tokens (default: 4000)

3. After configuration, you'll see the GitGPT welcome interface!

```bash
🤖 Welcome to GitGPT!
------------------------

  Natural Language  - Use natural language to interact with Git
  commit           - Generate commit message and commit changes
  config           - Run configuration wizard
  cd <path>        - Change working directory
  exit             - Quit the application
```

## 💡 Usage Examples

### Natural Language Git Interaction

GitGPT supports two types of Git operations:

#### 1. Query Operations

For retrieving repository information without modifying repository state:

```bash
> What files were modified in the last week?

🔄 Executing: git log --name-status --since=2024-11-01
✅ Files modified in the last week:
- `README.md`
- `README_EN.md`
- `.goreleaser.yml`
- `cmd/main.go` (renamed to `cmd/gitchat/main.go`)
```

```bash
> Write a 100-word daily report based on the last commit

🔄 Executing: git log -p -1
✅ Today's work focused on enhancing Git repository validation functionality. I added new code in chat_agent.go
to verify if the current directory is a Git repository before executing chat functionality. If not, it returns
a custom NotGitRepoError. Additionally, I cleaned up the responses to ensure proper formatting. These changes
improve system robustness by preventing unnecessary operations in non-Git repository environments.
```

#### 2. Modification Operations

Can execute operations that change repository state, with confirmation before execution:

```bash
> I will develop a new login feature

ℹ️ Command1: git checkout -b feature/login-functionality
ℹ️ Purpose: Create a new branch for login feature development.
⚠️ Impact: This will create and switch to a new branch named 'feature/login-functionality' to develop without affecting the main branch.

Do you want to execute these commands? (y/n): y
✅ Executed: git checkout -b feature/login-functionality
Switched to a new branch 'feature/login-functionality'
```

### Smart Commit Message Generation

When you want to commit code changes:

```bash
> commit
```

GitGPT will analyze your changes and suggest appropriate commit messages:

```bash
📄 Staged Files:
------------------------
📝 internal/agent/commit_agent.go (16+/18-)

📝 Change Summary
------------------------
Enhanced the PrepareCommit function, added valid Git repository checking and refactored response handling.
Introduced a new error type for non-Git repositories and modified return type to include structured CommitResponse.
Improved logging and error handling in generateSuggestions function for clearer and more reliable suggestion generation.

💡 Suggested Commit Messages
------------------------

1) feat(agent): Add valid Git repository check
2) refactor(agent): Update response handling in PrepareCommit
3) fix(agent): Improve error handling in suggestion generation

Select a message (1-3), enter 'r' to regenerate, 'c' to cancel, or 'm' for manual input: 1
✅ Successfully committed changes with message: feat(agent): Add valid Git repository check
```

## 📬 Contact & Support

- Report issues or suggest features on the [Issues](https://github.com/go-coders/git_gpt/issues) page
- If you find it useful, please give us a Star!

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

Built with ❤️ by [Go Coders](https://github.com/go-coders)
