package agent

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-coders/git_gpt/internal/agent/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// BaseAgentTestSuite provides common test functionality
type BaseAgentTestSuite struct {
	suite.Suite
	ctx     context.Context
	git     *mocks.GitExecutor    // 更新类型
	llm     *mocks.LLMClient      // 更新类型
	display *mocks.DisplayManager // 更新类型
	logger  *mocks.Logger         // 更新类型
	input   *bytes.Buffer
}

func (s *BaseAgentTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.git = new(mocks.GitExecutor)
	s.llm = new(mocks.LLMClient)
	s.display = new(mocks.DisplayManager)
	s.logger = new(mocks.Logger)
	s.input = new(bytes.Buffer)

}

func (s *BaseAgentTestSuite) newTestAgent() (*BaseAgent, error) {
	config := AgentConfig{
		Git:     s.git,
		LLM:     s.llm,
		Display: s.display,
		Logger:  s.logger,
		Reader:  s.input,
	}

	return NewBaseAgent(config)
}

func TestBaseAgent(t *testing.T) {
	suite.Run(t, new(BaseAgentTestSuite))
}

// Test configuration validation
func (s *BaseAgentTestSuite) TestNewBaseAgent_ConfigValidation() {
	testCases := []struct {
		name        string
		config      AgentConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing git executor",
			config:      AgentConfig{LLM: s.llm, Display: s.display, Logger: s.logger, Reader: s.input},
			expectError: true,
			errorMsg:    "git executor is required",
		},
		{
			name:        "missing LLM client",
			config:      AgentConfig{Git: s.git, Display: s.display, Logger: s.logger, Reader: s.input},
			expectError: true,
			errorMsg:    "LLM client is required",
		},
		{
			name:        "missing display manager",
			config:      AgentConfig{Git: s.git, LLM: s.llm, Logger: s.logger, Reader: s.input},
			expectError: true,
			errorMsg:    "display manager is required",
		},
		{
			name:        "missing logger",
			config:      AgentConfig{Git: s.git, LLM: s.llm, Display: s.display, Reader: s.input},
			expectError: true,
			errorMsg:    "logger is required",
		},
		{
			name:        "valid config",
			config:      AgentConfig{Git: s.git, LLM: s.llm, Display: s.display, Logger: s.logger, Reader: s.input},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			agent, err := NewBaseAgent(tc.config)
			if tc.expectError {
				s.Assert().Error(err)
				s.Assert().Contains(err.Error(), tc.errorMsg)
				s.Assert().Nil(agent)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(agent)
			}
		})
	}
}

// Test command validation

// Test command execution
// Test command execution

// Test user confirmation prompt
func (s *BaseAgentTestSuite) TestPromptForConfirmation() {
	agent, err := s.newTestAgent()
	s.Require().NoError(err)

	testCases := []struct {
		name          string
		input         string
		expectedReply bool
		expectError   bool
	}{
		{
			name:          "user confirms",
			input:         "y\n",
			expectedReply: true,
		},
		{
			name:          "user declines",
			input:         "n\n",
			expectedReply: false,
		},
		{
			name:          "empty input",
			input:         "\n",
			expectedReply: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.input.Reset()
			s.input.WriteString(tc.input)

			confirmed, err := agent.promptForConfirmation("Confirm? ")

			if tc.expectError {
				s.Assert().Error(err)
			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(tc.expectedReply, confirmed)
			}
		})
	}
}

func (s *BaseAgentTestSuite) TestHandleCommandResults() {
	agent, err := s.newTestAgent()
	s.Require().NoError(err)

	testCases := []struct {
		name        string
		results     []CommandResult
		expectError bool
	}{
		{
			name: "successful results",
			results: []CommandResult{
				{
					Command: Command{Args: []string{"log", "-n", "1"}},
					Output:  "commit abc123",
				},
			},
			expectError: false,
		},
		{
			name: "command error",
			results: []CommandResult{
				{
					Command: Command{Args: []string{"log", "-n", "1"}},
					Error:   fmt.Errorf("command failed"),
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			display := new(mocks.DisplayManager)
			agent.display = display

			if !tc.expectError {
				// For successful case, only expect ShowSuccess
				display.On("ShowSuccess", fmt.Sprintf("Executed: git %s",
					strings.Join(tc.results[0].Command.Args, " "))).Return()
			} else {
				// For error case, only expect ShowError
				display.On("ShowError", mock.Anything).Return()
			}

			err := agent.handleCommandResults(s.ctx, tc.results)

			if tc.expectError {
				s.Assert().Error(err)
			} else {
				s.Assert().NoError(err)
			}

			display.AssertExpectations(s.T())
		})
	}
}

// ... existing code ...

// Test JSON response cleaning
func (s *BaseAgentTestSuite) TestCleanJSONResponse() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean json",
			input:    `{"type": "test"}`,
			expected: `{"type": "test"}`,
		},
		{
			name:     "json with markdown",
			input:    "```json\n{\"type\": \"test\"}\n```",
			expected: `{"type": "test"}`,
		},
		{
			name:     "json with whitespace",
			input:    "  {\"type\": \"test\"}  ",
			expected: `{"type": "test"}`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := cleanJSONResponse(tc.input)
			s.Assert().Equal(tc.expected, result)
		})
	}
}
