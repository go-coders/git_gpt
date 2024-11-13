package agent

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ChatAgentTestSuite struct {
	BaseAgentTestSuite
	agent *ChatAgent
}

func (s *ChatAgentTestSuite) SetupTest() {
	s.BaseAgentTestSuite.SetupTest()

	s.llm.On("SetSystemMessage", mock.Anything).Return()
	s.llm.On("ClearHistory").Return()

	config := AgentConfig{
		Git:     s.git,
		LLM:     s.llm,
		Display: s.display,
		Logger:  s.logger,
		Reader:  s.input,
	}

	agent, err := NewChatAgent(config)
	s.Require().NoError(err)
	s.agent = agent
}

func TestChatAgent(t *testing.T) {
	suite.Run(t, new(ChatAgentTestSuite))
}

func (s *ChatAgentTestSuite) TestChat_NotGitRepo() {
	s.git.On("IsGitRepository", s.ctx).Return(false)

	err := s.agent.Chat(s.ctx, "show last commit")
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "not a git repository")
}

func (s *ChatAgentTestSuite) TestChat_QueryCommand() {
	// Mock repository check
	s.git.On("IsGitRepository", s.ctx).Return(true)

	// Mock LLM response for command generation
	commandResponse := Response{
		Type:        "execute",
		CommandType: CommandTypeQuery,
		Commands: []Command{
			{
				Type:    CommandTypeQuery,
				Args:    []string{"log", "-n", "1"},
				Purpose: "Show last commit",
			},
		},
	}
	commandJSON, _ := json.Marshal(commandResponse)

	s.llm.On("Chat", s.ctx, mock.Anything).
		Return(string(commandJSON), nil).Once()

	// Mock git execution - 修改这里
	s.git.On("Execute", s.ctx, "log", "-n", "1").
		Return("commit abc123\nAuthor: Test\nDate: 2024\n\nTest commit", nil)

	// Mock LLM summary
	s.llm.On("Chat", s.ctx, mock.Anything).
		Return("Last commit was by Test on 2024", nil).Once()

	// Display expectations
	s.display.On("StartSpinner", mock.Anything).Return()
	s.display.On("StopSpinner").Return()
	s.display.On("ShowCommand", mock.Anything).Return()
	s.display.On("ShowSuccess", "Last commit was by Test on 2024").Return()

	err := s.agent.Chat(s.ctx, "show last commit")
	s.Assert().NoError(err)
}

func (s *ChatAgentTestSuite) TestChat_ModifyCommand() {
	// Mock repository check
	s.git.On("IsGitRepository", s.ctx).Return(true)

	// Mock LLM response for command generation
	commandResponse := Response{
		Type:        "execute",
		CommandType: CommandTypeModify,
		Commands: []Command{
			{
				Type:    CommandTypeModify,
				Args:    []string{"reset", "--hard", "HEAD~1"},
				Purpose: "Reset to previous commit",
				Impact:  "This will discard the last commit",
			},
		},
	}
	commandJSON, _ := json.Marshal(commandResponse)

	s.llm.On("Chat", s.ctx, mock.Anything).
		Return(string(commandJSON), nil)

	// Display expectations
	s.display.On("StartSpinner", mock.Anything).Return()
	s.display.On("StopSpinner").Return()
	s.display.On("ShowInfo", mock.Anything).Return()
	s.display.On("ShowWarning", mock.Anything).Return()

	// Simulate user declining the operation
	s.input.WriteString("n\n")

	err := s.agent.Chat(s.ctx, "reset to previous commit")
	s.Assert().NoError(err)
}

func (s *ChatAgentTestSuite) TestChat_DirectAnswer() {
	// Mock repository check
	s.git.On("IsGitRepository", s.ctx).Return(true)

	// Mock LLM response with direct answer
	response := Response{
		Type:    "answer",
		Content: "This is a direct answer",
	}
	responseJSON, _ := json.Marshal(response)

	s.llm.On("Chat", s.ctx, mock.Anything).
		Return(string(responseJSON), nil)

	// Display expectations
	s.display.On("StartSpinner", mock.Anything).Return()
	s.display.On("StopSpinner").Return()
	s.display.On("ShowSuccess", "This is a direct answer").Return()

	err := s.agent.Chat(s.ctx, "what is git?")
	s.Assert().NoError(err)
}

func (s *ChatAgentTestSuite) TestGetCommandResponse() {
	// Mock LLM response
	commandResponse := Response{
		Type:        "execute",
		CommandType: CommandTypeQuery,
		Commands: []Command{
			{
				Type:    CommandTypeQuery,
				Args:    []string{"log", "-n", "1"},
				Purpose: "Show last commit",
			},
		},
	}
	commandJSON, _ := json.Marshal(commandResponse)

	s.llm.On("Chat", s.ctx, mock.Anything).Return(string(commandJSON), nil)
	s.display.On("StartSpinner", mock.Anything).Return()
	s.display.On("StopSpinner").Return()

	response, err := s.agent.getCommandResponse(s.ctx, "show last commit")
	s.Assert().NoError(err)
	s.Assert().Equal(commandResponse, response)
}

func (s *ChatAgentTestSuite) TestHandleResponse_Answer() {
	response := Response{
		Type:    "answer",
		Content: "This is a direct answer",
	}

	s.display.On("ShowSuccess", "This is a direct answer").Return()

	err := s.agent.handleResponse(s.ctx, "what is git?", response)
	s.Assert().NoError(err)
}

// Test handleModificationCommands with user rejection
func (s *ChatAgentTestSuite) TestHandleModificationCommands_Rejected() {
	commands := []Command{
		{
			Type:    CommandTypeModify,
			Args:    []string{"reset", "--hard", "HEAD~1"},
			Purpose: "Reset to previous commit",
			Impact:  "This will discard the last commit",
		},
	}

	s.display.On("ShowWarning", mock.Anything).Return()
	s.display.On("ShowInfo", mock.Anything).Return()

	// Simulate user rejecting the operation
	s.input.WriteString("n\n")

	err := s.agent.handleModificationCommands(s.ctx, commands)
	s.Assert().NoError(err)
}
