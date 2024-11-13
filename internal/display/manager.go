// internal/display/manager.go
package display

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

type Section struct {
	Title    string
	Content  string
	Icon     string
	Divider  string
	Numbered bool
}

type DisplayImpl struct {
	spinner   *spinner.Spinner
	formatter *ColorFormatter
	mu        sync.Mutex // Protect spinner operations
}

func NewManager(version string) *DisplayImpl {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Color("cyan")
	return &DisplayImpl{
		spinner:   s,
		formatter: NewColorFormatter(version),
	}
}

func (m *DisplayImpl) ShowSection(title, content string, opts map[string]string) {
	divider := opts["divider"]
	if divider == "" {
		divider = "------------------------"
	}

	displayTitle := title
	if icon := opts["icon"]; icon != "" {
		displayTitle = fmt.Sprintf("%s %s", icon, title)
	}

	fmt.Printf("\n%s\n%s\n", m.formatter.FormatSectionTitle(displayTitle), divider)
	if content != "" {
		fmt.Println(m.formatter.FormatSectionContent(content))
	}
}

func (m *DisplayImpl) ShowNumberedList(items [][2]string) {
	for i, item := range items {
		// Pass both index and content to FormatListItem
		fmt.Printf("%s\n", m.formatter.FormatListItem(i+1, item[0]))
		if item[1] != "" {
			fmt.Printf("   %s\n", m.formatter.FormatListDescription(item[1]))
		}
	}
}

// ShowWelcome displays the welcome message
func (m *DisplayImpl) ShowWelcome() {
	fmt.Println(m.formatter.FormatWelcome())
}

// Required methods for agent.DisplayManager interface
func (m *DisplayImpl) ShowPrompt(pwd, branch string) {
	fmt.Print(m.formatter.FormatPrompt(pwd, branch))
}

func (m *DisplayImpl) ShowSuccess(message string) {
	m.stopSpinnerIfActive()
	fmt.Println(m.formatter.FormatSuccess(message))
}

func (m *DisplayImpl) ShowError(message string) {
	m.stopSpinnerIfActive()
	fmt.Println(m.formatter.FormatError(message))
}

func (m *DisplayImpl) ShowInfo(message string) {
	m.stopSpinnerIfActive()
	fmt.Println(m.formatter.FormatInfo(message))
}

func (m *DisplayImpl) ShowWarning(message string) {
	m.stopSpinnerIfActive()
	fmt.Println(m.formatter.FormatWarning(message))
}

func (m *DisplayImpl) ShowCommand(command string) {
	m.stopSpinnerIfActive()
	fmt.Println(m.formatter.FormatCommand(command))
}

func (m *DisplayImpl) StartSpinner(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.spinner.Suffix = fmt.Sprintf(" %s", message)
	if !m.spinner.Active() {
		m.spinner.Start()
	}
}

func (m *DisplayImpl) StopSpinner() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.spinner.Active() {
		m.spinner.Stop()
	}
}

func (m *DisplayImpl) stopSpinnerIfActive() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.spinner.Active() {
		m.spinner.Stop()
	}
}

type ColorFormatter struct {
	success *color.Color
	error   *color.Color
	info    *color.Color
	warning *color.Color
	prompt  *color.Color
	title   *color.Color
	command *color.Color
	version string // Add version field
}

func NewColorFormatter(version string) *ColorFormatter {
	return &ColorFormatter{
		success: color.New(color.FgGreen, color.Bold),
		error:   color.New(color.FgRed, color.Bold),
		info:    color.New(color.FgCyan),
		warning: color.New(color.FgYellow, color.Bold),
		prompt:  color.New(color.FgYellow, color.Bold),
		title:   color.New(color.FgMagenta, color.Bold),
		command: color.New(color.FgBlue),
		version: version,
	}
}

func (f *ColorFormatter) FormatPrompt(pwd, branch string) string {
	return fmt.Sprintf("\n📂 %s [%s]\n> ", pwd, branch)
}

func (f *ColorFormatter) FormatCommand(command string) string {
	return f.command.Sprintf("🔄 Executing: %s", command)
}

func (f *ColorFormatter) FormatSuccess(message string) string {
	return f.success.Sprintf("✅ %s", message)
}

func (f *ColorFormatter) FormatError(message string) string {
	return f.error.Sprintf("❌ %s", message)
}

func (f *ColorFormatter) FormatInfo(message string) string {
	return f.info.Sprintf("ℹ️  %s", message)
}

func (f *ColorFormatter) FormatWarning(message string) string {
	return f.warning.Sprintf("⚠️  %s", message)
}

func (f *ColorFormatter) FormatWelcome() string {
	var b strings.Builder

	// Title
	b.WriteString(f.title.Sprintf("\n🤖 Welcome to GitChat v%s\n", f.version))
	b.WriteString(f.title.Sprintf("------------------------\n\n"))

	// Command list with descriptions in two languages
	commands := []struct {
		cmd, descEn, descZh string
	}{
		{
			cmd:    "Natural Language",
			descEn: "Use natural language to interact with Git",
			descZh: "使用自然语言与Git交互",
		},
		{
			cmd:    "commit",
			descEn: "Generate commit message and commit changes",
			descZh: "生成提交消息并提交更改",
		},
		{
			cmd:    "config",
			descEn: "Run configuration wizard",
			descZh: "运行配置向导",
		},
		{
			cmd:    "cd <path>",
			descEn: "Change working directory",
			descZh: "更改工作目录",
		},
		{
			cmd:    "exit",
			descEn: "Quit the application",
			descZh: "退出应用程序",
		},
	}

	// Calculate maximum width for command column
	maxCmdWidth := 0
	for _, cmd := range commands {
		if len(cmd.cmd) > maxCmdWidth {
			maxCmdWidth = len(cmd.cmd)
		}
	}

	// Add padding for better visual separation
	maxCmdWidth += 2 // Add some extra space

	for _, cmd := range commands {
		// Command with padding
		paddedCmd := cmd.cmd + strings.Repeat(" ", maxCmdWidth-len(cmd.cmd))

		// Write command and English description
		b.WriteString(f.command.Sprintf("  %s", paddedCmd))
		b.WriteString(f.info.Sprintf("- %s\n", cmd.descEn))

		// Write Chinese description with same padding
		padding := strings.Repeat(" ", maxCmdWidth+2)
		b.WriteString(f.info.Sprintf("  %s  %s\n", padding, cmd.descZh))
	}

	return b.String()
}

// DisplayTest runs a test of all display features
func (m *DisplayImpl) DisplayTest() {
	m.ShowWelcome()
	m.ShowPrompt("/home/user", "main")
	m.ShowSuccess("This is a success message")
	m.ShowError("This is an error message")
	m.ShowInfo("This is an info message")
	m.ShowWarning("This is a warning message")

	m.StartSpinner("Processing...")
	time.Sleep(2 * time.Second)
	m.StopSpinner()
}

// Add to ColorFormatter
type ListItem struct {
	Content     string
	Description string
}

func (f *ColorFormatter) FormatSectionTitle(title string) string {
	return f.title.Sprint(title)
}

func (f *ColorFormatter) FormatSectionContent(content string) string {
	return f.info.Sprint(content)
}

func (f *ColorFormatter) FormatListItem(number int, content string) string {
	return fmt.Sprintf("%d) %s\n", number, f.command.Sprint(content))
}

func (f *ColorFormatter) FormatListDescription(description string) string {
	return fmt.Sprintf("   %s\n", f.info.Sprint(description))
}
