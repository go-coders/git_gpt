// internal/app/config_wizard.go
package app

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-coders/gitchat/internal/config"
	"github.com/go-coders/gitchat/pkg/apierrors"
)

type ConfigWizard struct {
	config *config.Config
	reader *bufio.Reader
}

func NewConfigWizard(cfg *config.Config) *ConfigWizard {
	return &ConfigWizard{
		config: cfg,
		reader: bufio.NewReader(os.Stdin),
	}
}

type configPrompt struct {
	label     string
	current   string
	validator func(string) error
	setter    func(string) error
}

func (w *ConfigWizard) Run() error {
	fmt.Println("\n🔧 LLM Configuration")
	fmt.Println("------------------------")

	var validValue bool
	defer func(cf config.Config) {
		// reset config if invalid value
		if !validValue {
			*w.config = cf
		}
	}(*w.config)

	prompts := w.getPrompts()
	for _, p := range prompts {
		if err := w.handlePrompt(p); err != nil {
			return err
		}
	}
	if err := w.config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	validValue = true

	w.showSummary()
	return nil
}

func (w *ConfigWizard) getPrompts() []configPrompt {
	return []configPrompt{
		{
			label:   "LLM API Key",
			current: maskAPIKey(w.config.LLM.APIKey),
			validator: func(s string) error {
				if s == "" && w.config.LLM.APIKey == "" {
					return apierrors.NewApiKeyError()
				}
				return nil
			},
			setter: func(s string) error {
				if s != "" {
					w.config.LLM.APIKey = s
				}
				return nil
			},
		},
		{
			label:   "LLM Model",
			current: w.config.LLM.Model,
			setter: func(s string) error {
				if s != "" {
					w.config.LLM.Model = s
				}
				return nil
			},
		},
		{
			label:   "LLM API Base URL",
			current: w.config.LLM.BaseURL,
			setter: func(s string) error {
				if s != "" {
					w.config.LLM.BaseURL = s
				}
				return nil
			},
		},
		{
			label:   "Max Tokens",
			current: strconv.Itoa(w.config.LLM.MaxTokens),
			validator: func(s string) error {
				if s != "" {
					if _, err := strconv.Atoi(s); err != nil {
						return fmt.Errorf("invalid number format")
					}
				}
				return nil
			},
			setter: func(s string) error {
				if s != "" {
					tokens, _ := strconv.Atoi(s)
					w.config.LLM.MaxTokens = tokens
				}
				return nil
			},
		},
	}
}

func (w *ConfigWizard) handlePrompt(p configPrompt) error {
	var prompt string
	if p.current != "" {
		prompt = fmt.Sprintf("Enter %s (current: %s, press Enter to keep current): ", p.label, p.current)
	} else {
		prompt = fmt.Sprintf("Enter %s: ", p.label)
	}

	fmt.Print(prompt)
	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if p.validator != nil {
		if err := p.validator(input); err != nil {
			return fmt.Errorf("%s: %w", p.label, err)
		}
	}

	return p.setter(input)
}

func (w *ConfigWizard) showSummary() {
	fmt.Println("\n📝 Configuration Summary:")
	fmt.Printf("API Key: %s\n", maskAPIKey(w.config.LLM.APIKey))
	fmt.Printf("Model: %s\n", w.config.LLM.Model)
	fmt.Printf("API Base URL: %s\n", w.config.LLM.BaseURL)
	fmt.Printf("Max Tokens: %d\n", w.config.LLM.MaxTokens)
	fmt.Printf("\n✅ Configuration saved to: %s\n", w.config.ConfigPath)
}

func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	if len(apiKey) < 6 {
		return apiKey
	}
	return "***" + apiKey[len(apiKey)-6:]
}
