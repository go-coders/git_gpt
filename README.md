# GitGPT

简体中文 | [English](README_EN.md)

GitGPT 是一个革新性的命令行工具，它将 GPT 大语言模型与 Git 完美结合，让你能用自然语言与 Git 进行对话式交互。无需记忆复杂的 Git 命令，你可以用日常对话的方式执行 Git 操作，比如"帮我创建一个登录功能的分支"或"查看上周的代码改动"。它不仅能理解你的意图，还会在执行关键操作前提供清晰的解释和确认，让 Git 操作变得更加智能、安全和高效。

<div align="center">

[![Release](https://img.shields.io/github/v/release/go-coders/git_gpt)](https://github.com/go-coders/git_gpt/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-coders/git_gpt)](https://goreportcard.com/report/github.com/go-coders/git_gpt)
[![Tests](https://github.com/go-coders/git_gpt/actions/workflows/test.yml/badge.svg)](https://github.com/go-coders/git_gpt/actions/workflows/test.yml)
[![Coverage Status](https://codecov.io/gh/go-coders/git_gpt/branch/main/graph/badge.svg)](https://codecov.io/gh/go-coders/git_gpt)
[![License](https://img.shields.io/github/license/go-coders/git_gpt)](LICENSE)

</div>

## 📦 安装

### 使用 Go Install (需要 Go 1.20+)

```bash
go install github.com/go-coders/git_gpt/cmd/ggpt@latest
```

### 下载预编译二进制文件

前往 [Releases](https://github.com/go-coders/git_gpt/releases) 页面，下载适用于您操作系统的可执行文件，并将其添加到系统的 PATH 中。

## 🚀 快速开始

1. 安装完成后，在终端中运行：

   ```bash
   ggpt
   ```

2. 首次运行时，会启动配置向导。你需要提供：

   - OpenAI API 密钥
   - 模型选择（默认：gpt-4o）
   - API 基础 URL（默认：https://api.openai.com/v1）
   - 最大 token 数（默认：4000）

3. 配置完成后，即可以看到 GitGPT 的欢迎界面！

```bash
🤖 Welcome to GitGPT!
------------------------

  Natural Language  - Use natural language to interact with Git
                     使用自然语言与Git交互
  commit           - Generate commit message and commit changes
                     生成提交消息并提交更改
  config           - Run configuration wizard
                     运行配置向导
  cd <path>        - Change working directory
                     更改工作目录
  exit             - Quit the application
                     退出应用程序
```

## 💡 使用示例

### 自然语言 Git 交互

GitGPT 支持两种类型的 Git 操作：

#### 1. 查询操作

用于获取仓库信息，不会修改仓库状态：

```bash
> 最近一周主要修改了哪些文件

🔄 Executing: git log --name-status --since=2024-11-01
✅ 最近一周主要修改的文件有：
- `README.md`
- `README_EN.md`
- `.goreleaser.yml`
- `cmd/main.go`（重命名为 `cmd/gitchat/main.go`）
```

```bash
> 根据最后一次提交的具体内容写一篇 100 字的日报

🔄 执行中: git log -p -1
✅ 今天的工作主要集中在增强 Git 仓库的检查功能。我在 chat_agent.go 文件中新增了一段代码，
用于在执行聊天功能前验证当前目录是否为 Git 仓库。如果不是，则返回一个自定义错误 NotGitRepoError。
此外，我还对响应进行了清理，以确保格式正确。这些改动提高了系统的健壮性，避免了在非 Git 仓库环境下执行不必要的操作。
```

#### 2. 修改操作

可以执行会改变仓库状态的操作，执行前会请求确认：

```bash
> 我将开发一个登录的新功能

ℹ️ Command1: git checkout -b feature/login-functionality
ℹ️ Purpose: 创建一个新的分支来开发登录功能。
⚠️ Impact: 这将创建并切换到一个名为 'feature/login-functionality' 的新分支，以便在不影响主分支的情况下进行开发。

Do you want to execute these commands? (y/n): y
✅ Executed: git checkout -b feature/login-functionality
Switched to a new branch 'feature/login-functionality'
```

### 智能提交消息生成

当你想要提交代码更改时：

```bash
> commit
```

GitGPT 将分析你的更改并建议合适的提交信息：

```bash
📄 已暂存的文件:
------------------------
📝 internal/agent/commit_agent.go (16+/18-)

📝 变更摘要
------------------------
增强了 PrepareCommit 函数，增加了对有效 Git 仓库的检查并重构了响应处理。
引入了一个新的错误类型用于非 Git 仓库，并修改了返回类型以包含结构化的 CommitResponse。
改进了 generateSuggestions 函数中的日志和错误处理，确保了更清晰和可靠的建议生成。

💡 建议的提交消息
------------------------

1) feat(agent): 添加有效 Git 仓库的检查
2) refactor(agent): 更新 PrepareCommit 中的响应处理
3) fix(agent): 改进建议生成中的错误处理

请选择一个消息 (1-3)，输入 'r' 重新生成，输入 'c' 取消，或输入 'm' 手动输入: 1
✅ 已成功提交更改，提交消息: feat(agent): 添加有效 Git 仓库的检查
```

## 📬 联系与支持

- 在 [Issues](https://github.com/go-coders/git_gpt/issues) 页面报告问题或提出功能建议
- 如果觉得有用，请给我们一个 Star！

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。

---

由 [Go Coders](https://github.com/go-coders) 用 ❤️ 构建
