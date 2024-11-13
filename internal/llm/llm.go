package llm

import (
	"context"
	"errors"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

var (
	ErrExceedMaxTokens = errors.New("message tokens exceed maximum limit")
	ErrEmptyMessage    = errors.New("message content cannot be empty")
	ErrInvalidAPIKey   = errors.New("invalid API key")
)

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"

	responseTokenReserve = 300
)

type (
	Role string

	Logger interface {
		Debug(format string, args ...interface{})
		Error(format string, args ...interface{})
	}

	Message struct {
		Role    Role
		Content string
	}

	Config struct {
		APIKey        string
		BaseURL       string
		EnableHistory bool
		MaxMessages   int // 0 means no limit
		MaxTokens     int // 0 means no limit
		Model         string
		Temperature   float32
	}

	Client struct {
		client         *openai.Client
		config         Config
		messageHistory []Message
		tokenizer      *tiktoken.Tiktoken
		systemMessage  string
	}
)

func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, ErrInvalidAPIKey
	}

	if config.Temperature == 0 {
		config.Temperature = 0.1
	}

	cfg := openai.DefaultConfig(config.APIKey)
	cfg.BaseURL = config.BaseURL

	client := &Client{
		client:         openai.NewClientWithConfig(cfg),
		config:         config,
		messageHistory: make([]Message, 0),
	}

	tkm, err := client.getTokenEncoder()
	if err != nil {
		return nil, err
	}
	client.tokenizer = tkm

	return client, nil
}

func (c *Client) Chat(ctx context.Context, content string) (string, error) {
	if content == "" {
		return "", ErrEmptyMessage
	}

	messages := c.prepareMessages(Message{Role: RoleUser, Content: content})
	response, err := c.sendRequest(ctx, messages)
	if err != nil {
		return "", err
	}

	if c.config.EnableHistory {
		c.updateHistory(messages, response)
	}

	return response, nil
}

func (c *Client) prepareMessages(newMessage Message) []Message {
	// Calculate available tokens
	totalAvailable := c.calculateAvailableTokens()
	newMsgTokens := c.countTokens([]Message{newMessage})

	// Build initial messages list with system message
	var messages []Message
	var systemTokens int
	if c.systemMessage != "" {
		sysMsg := Message{Role: RoleSystem, Content: c.systemMessage}
		systemTokens = c.countTokens([]Message{sysMsg})
		messages = append(messages, sysMsg)
	}

	// Handle no history mode
	if !c.config.EnableHistory {
		if c.config.MaxTokens > 0 && newMsgTokens+systemTokens > totalAvailable {
			return []Message{c.truncateMessage(newMessage, totalAvailable)}
		}
		return append(messages, newMessage)
	}

	// Handle history mode
	if c.config.MaxTokens > 0 {
		return c.buildHistoryMessages(messages, newMessage)
	}

	if c.config.MaxMessages > 0 {
		return c.buildMessageCountLimited(messages, newMessage)
	}

	return append(messages, append(c.messageHistory, newMessage)...)
}

func (c *Client) buildHistoryMessages(messages []Message, newMessage Message) []Message {
	totalAvailable := c.config.MaxTokens - responseTokenReserve
	newMsgTokens := c.countTokens([]Message{newMessage})

	// If new message is too long, truncate and return
	if newMsgTokens > totalAvailable {
		return []Message{c.truncateMessage(newMessage, totalAvailable)}
	}

	// Check system message + new message fit
	systemTokens := 0
	if len(messages) > 0 {
		systemTokens = c.countTokens(messages)
		if systemTokens+newMsgTokens > totalAvailable {
			return []Message{c.truncateMessage(newMessage, totalAvailable)}
		}
	}

	// Add history messages that fit
	var history []Message
	remainingTokens := totalAvailable - newMsgTokens - systemTokens

	for i := len(c.messageHistory) - 1; i >= 0; i-- {
		msg := c.messageHistory[i]
		msgTokens := c.countTokens([]Message{msg})
		if msgTokens > remainingTokens {
			break
		}
		history = append([]Message{msg}, history...)
		remainingTokens -= msgTokens
	}

	// If no history fits, return just new message
	if len(history) == 0 {
		return []Message{newMessage}
	}

	result := messages
	result = append(result, history...)
	return append(result, newMessage)
}

func (c *Client) truncateMessage(msg Message, maxTokens int) Message {
	tokens := c.tokenizer.Encode(msg.Content, nil, nil)
	if len(tokens) <= maxTokens {
		return msg
	}

	const suffix = "..."
	suffixTokens := len(c.tokenizer.Encode(suffix, nil, nil))
	availableTokens := maxTokens - suffixTokens

	truncated := c.tokenizer.Decode(tokens[:availableTokens])

	// Find best breaking point
	for _, breakPoint := range []struct {
		sep string
		add string
	}{
		{". ", ". ..."},
		{"\n", "\n..."},
		{" ", "..."},
	} {
		if idx := strings.LastIndex(truncated, breakPoint.sep); idx > len(truncated)/2 {
			return Message{
				Role:    msg.Role,
				Content: truncated[:idx] + breakPoint.add,
			}
		}
	}

	return Message{
		Role:    msg.Role,
		Content: truncated + suffix,
	}
}

// Helper methods...
func (c *Client) countTokens(messages []Message) int {
	tokens := 0
	for _, msg := range messages {
		tokens += len(c.tokenizer.Encode(msg.Content, nil, nil)) + 4
	}
	return tokens
}

func (c *Client) calculateAvailableTokens() int {
	if c.config.MaxTokens <= 0 {
		return 0
	}
	return c.config.MaxTokens - responseTokenReserve
}

func (c *Client) buildMessageCountLimited(messages []Message, newMessage Message) []Message {
	if c.config.MaxMessages <= 0 {
		return append(messages, append(c.messageHistory, newMessage)...)
	}

	// Calculate how many historical messages we can include
	maxHistory := c.config.MaxMessages - 1 // -1 for new message
	if len(messages) > 0 {                 // If we have system message
		maxHistory-- // -1 for system message
	}

	if maxHistory <= 0 {
		// If no room for history, just return new message
		return []Message{newMessage}
	}

	// Take most recent history
	historyStart := 0
	if len(c.messageHistory) > maxHistory {
		historyStart = len(c.messageHistory) - maxHistory
	}

	result := messages // Contains system message if present
	result = append(result, c.messageHistory[historyStart:]...)
	result = append(result, newMessage)

	return result
}

func (c *Client) getTokenEncoder() (*tiktoken.Tiktoken, error) {
	tkm, err := tiktoken.EncodingForModel(c.config.Model)
	if err != nil {
		// Fallback to cl100k_base encoding (used by GPT-3.5 and GPT-4)
		tkm, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return nil, err
		}
	}
	return tkm, nil
}

func (c *Client) sendRequest(ctx context.Context, messages []Message) (string, error) {
	openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    openAIMessages,
		Temperature: c.config.Temperature,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response received")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *Client) updateHistory(messages []Message, response string) {
	if !c.config.EnableHistory {
		return
	}

	// Skip system message when updating history
	startIdx := 0
	if len(messages) > 0 && messages[0].Role == RoleSystem {
		startIdx = 1
	}

	c.messageHistory = append([]Message{}, messages[startIdx:]...)
	c.messageHistory = append(c.messageHistory, Message{
		Role:    RoleAssistant,
		Content: response,
	})
}

func (c *Client) ClearHistory() {
	c.messageHistory = nil
}

func (c *Client) SetSystemMessage(message string) {
	c.systemMessage = message
}
