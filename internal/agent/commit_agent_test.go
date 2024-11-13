package agent

import (
	"testing"

	"github.com/go-coders/git_gpt/internal/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CommitAgentTestSuite struct {
	BaseAgentTestSuite
	agent *CommitAgent
}

func (s *CommitAgentTestSuite) SetupTest() {
	s.BaseAgentTestSuite.SetupTest()

	// Set up common mocks
	s.display.On("StartSpinner", mock.Anything).Return()
	s.display.On("StopSpinner").Return()
	s.logger.On("Debug", mock.Anything).Return()
	s.logger.On("Debug", mock.Anything, mock.Anything).Return()
	s.logger.On("Info", mock.Anything).Return()
	s.logger.On("Error", mock.Anything).Return()

	config := AgentConfig{
		Git:     s.git,
		LLM:     s.llm,
		Display: s.display,
		Logger:  s.logger,
		Reader:  s.input,
	}

	agent, err := NewCommitAgent(config)
	s.Require().NoError(err)
	s.agent = agent
}

func TestCommitAgent(t *testing.T) {
	suite.Run(t, new(CommitAgentTestSuite))
}

// Test helper functions
func (s *CommitAgentTestSuite) TestHelperFunctions() {
	// Test hasNoChanges
	status := &commitStatus{
		staged:   []common.FileChange{},
		unstaged: []common.FileChange{},
	}
	s.Assert().True(s.agent.hasNoChanges(status))

	// Test hasOnlyUnstagedChanges
	status.unstaged = []common.FileChange{{Path: "test.txt"}}
	s.Assert().True(s.agent.hasOnlyUnstagedChanges(status))

	// Test categorizeChanges
	changes := []common.FileChange{
		{Path: "modified.txt", Status: "modified"},
		{Path: "new.txt", Status: "untracked"},
	}
	modified, untracked := s.agent.categorizeChanges(changes)
	s.Assert().Len(modified, 1)
	s.Assert().Len(untracked, 1)
}

// Test getStatusSymbol
func (s *CommitAgentTestSuite) TestGetStatusSymbol() {
	testCases := map[string]string{
		"modified":  "📝",
		"added":     "➕",
		"deleted":   "➖",
		"renamed":   "📋",
		"copied":    "📑",
		"untracked": "❓",
		"unknown":   "•",
	}

	for status, expected := range testCases {
		s.Assert().Equal(expected, getStatusSymbol(status))
	}
}

func (s *CommitAgentTestSuite) TestHandleCommit_NoChanges() {
	// Mock git status with no changes
	s.git.On("GetStatus", s.ctx).Return(
		[]common.FileChange{},
		[]common.FileChange{},
		nil,
	).Once()

	// Expect info message
	s.display.On("ShowInfo", "No changes to commit").Return().Once()

	err := s.agent.HandleCommit(s.ctx)
	s.Assert().NoError(err)
}

// Test HandleCommit with staged changes
func (s *CommitAgentTestSuite) TestHandleCommit_StagedChanges() {
	stagedFiles := []common.FileChange{
		{
			Path:      "test1.txt",
			Status:    "modified",
			Additions: 5,
			Deletions: 2,
		},
	}

	// Mock status checks
	s.git.On("GetStatus", s.ctx).Return(stagedFiles, []common.FileChange{}, nil).Once()

	// Mock diff
	s.git.On("GetDiff", s.ctx, true).Return("test diff", nil).Once()

	// Mock LLM response with correct struct types
	llmResponse := `{
			"summary": "Test changes",
			"suggestions": [
					{"message": "feat: test commit 1", "description": "desc 1"},
					{"message": "fix: test commit 2", "description": "desc 2"},
					{"message": "docs: test commit 3", "description": "desc 3"}
			]
	}`
	s.llm.On("Chat", s.ctx, mock.Anything).Return(llmResponse, nil).Once()

	// Mock display calls
	s.display.On("StartSpinner", mock.Anything).Return()
	s.display.On("StopSpinner").Return()
	s.display.On("ShowSection", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
	s.display.On("ShowNumberedList", mock.Anything).Return().Times(2)

	// Mock user selecting first suggestion
	s.input.WriteString("1\n")

	// Mock git commit
	s.git.On("Commit", s.ctx, "feat: test commit 1").Return(nil).Once()
	s.display.On("ShowSuccess", mock.Anything).Return()

	err := s.agent.HandleCommit(s.ctx)
	s.Assert().NoError(err)
}

// Test HandleCommit with manual commit message
func (s *CommitAgentTestSuite) TestHandleCommit_ManualMessage() {
	stagedFiles := []common.FileChange{
		{Path: "test1.txt", Status: "modified"},
	}

	// Mock status checks
	s.git.On("GetStatus", s.ctx).Return(stagedFiles, []common.FileChange{}, nil).Once()

	// Mock diff
	s.git.On("GetDiff", s.ctx, true).Return("test diff", nil).Once()

	// Mock LLM response
	llmResponse := `{
			"summary": "Test changes",
			"suggestions": [
					{"message": "feat: test commit", "description": "desc"}
			]
	}`
	s.llm.On("Chat", s.ctx, mock.Anything).Return(llmResponse, nil).Once()

	// Mock display calls
	s.display.On("StartSpinner", mock.Anything).Return()
	s.display.On("StopSpinner").Return()
	s.display.On("ShowSection", mock.Anything, mock.Anything, mock.Anything).Return().Times(3)
	s.display.On("ShowNumberedList", mock.Anything).Return()

	// Mock user choosing manual input and entering message
	s.input.WriteString("m\n")
	s.input.WriteString("feat: manual commit message\n")

	// Mock git commit
	s.git.On("Commit", s.ctx, "feat: manual commit message").Return(nil).Once()
	s.display.On("ShowSuccess", mock.Anything).Return()

	err := s.agent.HandleCommit(s.ctx)
	s.Assert().NoError(err)
}

// Test HandleCommit with unstaged changes
func (s *CommitAgentTestSuite) TestHandleCommit_UnstagedChanges() {
	unstagedFiles := []common.FileChange{
		{Path: "test1.txt", Status: "modified"},
		{Path: "test2.txt", Status: "untracked"},
	}

	// Mock status checks - need to handle recursive call
	s.git.On("GetStatus", s.ctx).Return(
		[]common.FileChange{},
		unstagedFiles,
		nil,
	).Once()

	// Mock second status check after user declines
	s.git.On("GetStatus", s.ctx).Return(
		[]common.FileChange{},
		[]common.FileChange{},
		nil,
	).Once()

	// Mock display calls
	s.display.On("ShowSection", "Modified Files", "", mock.Anything).Return()
	s.display.On("ShowSection", "Untracked Files", "", mock.Anything).Return()
	s.display.On("ShowInfo", mock.Anything).Return().Times(3)        // Called for each file and cancel message
	s.display.On("ShowInfo", "No changes to commit").Return().Once() // Final status message

	// Mock user declining to stage
	s.input.WriteString("n\n")

	err := s.agent.HandleCommit(s.ctx)
	s.Assert().NoError(err)
}
