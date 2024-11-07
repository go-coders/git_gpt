// internal/llm/client.go
package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-coders/gitchat/pkg/apierrors"
	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client    *openai.Client
	model     string
	maxTokens int
}

const (
	RoleSystem = "system"
	RoleUser   = "user"
	RoleAssist = "assistant"
)

type ClientConfig struct {
	APIKey    string
	Model     string
	BaseURL   string
	MaxTokens int
}

func NewClient(cfg ClientConfig) (*Client, error) {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	return &Client{
		client:    openai.NewClientWithConfig(config),
		model:     cfg.Model,
		maxTokens: cfg.MaxTokens,
	}, nil
}

func (c *Client) Complete(ctx context.Context, messages []ConversationMessage) (string, error) {

	apiMsgs := convertToAPIMessages(messages)
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    apiMsgs,
		Temperature: 0.1,
	})

	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

type TokenLimitError struct {
	Current int
	Max     int
}

func (e *TokenLimitError) Error() string {
	return fmt.Sprintf("token limit exceeded: current %d, max %d", e.Current, e.Max)
}

func convertToAPIMessages(messages []ConversationMessage) []openai.ChatCompletionMessage {
	var apiMsgs []openai.ChatCompletionMessage
	for _, msg := range messages {
		apiMsgs = append(apiMsgs, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return apiMsgs
}

func (c *Client) CountTokens(messages []ConversationMessage) (int, error) {
	return c.numTokensFromMessages(messages)
}

type TokenExceededError struct {
	CurrentTokens int
	MaxTokens     int
}

func (e *TokenExceededError) Error() string {
	return fmt.Sprintf("token limit exceeded: current %d, max %d", e.CurrentTokens, e.MaxTokens)
}

type Config struct {
	APIKey    string
	Model     string
	BaseURL   string
	MaxTokens int
}

type ConversationMessage struct {
	Role    string
	Content string
	Tokens  int
	Origin  string
}

func (c *Client) numTokensFromMessages(messages []ConversationMessage) (int, error) {
	tkm, err := tiktoken.EncodingForModel(c.model)
	if err != nil {
		tkm, err = tiktoken.EncodingForModel("gpt-3.5-turbo")
		if err != nil {
			return 0, fmt.Errorf("failed to get encoding for model: %w", err)
		}
	}

	var numTokens int
	for _, message := range messages {
		numTokens += 3
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Content, nil, nil))
	}
	numTokens += 3
	return numTokens, nil
}

func (c *Client) MaxTokens() int {
	return c.maxTokens
}
func ValidateKey(apiKey, baseURL, model string) error {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	models, err := client.ListModels(ctx)

	if err != nil {
		return apierrors.NewAPIKeyOrUrlError(err)
	}

	for _, m := range models.Models {
		if model == m.ID {
			return nil
		}
	}

	return apierrors.NewInvalidModelError()

}
