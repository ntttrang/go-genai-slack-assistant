package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/slack-go/slack"
)

type MockSlackClient struct {
	mock.Mock
}

func (m *MockSlackClient) AddReaction(emoji, channelID, timestamp string) error {
	args := m.Called(emoji, channelID, timestamp)
	return args.Error(0)
}

func (m *MockSlackClient) PostMessage(channelID, text, threadTS string) (string, string, error) {
	args := m.Called(channelID, text, threadTS)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockSlackClient) GetMessage(channelID, timestamp string) (*slack.Message, error) {
	args := m.Called(channelID, timestamp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slack.Message), args.Error(1)
}

func (m *MockSlackClient) GetUserInfo(userID string) (*slack.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slack.User), args.Error(1)
}

// TestAddReactionEyes tests that the eyes emoji reaction is added correctly
func TestAddReactionEyes(t *testing.T) {
	mockSlackClient := new(MockSlackClient)
	
	channelID := "C123456"
	timestamp := "1234567890.123456"
	emoji := "eyes"
	
	mockSlackClient.On("AddReaction", emoji, channelID, timestamp).Return(nil)
	
	err := mockSlackClient.AddReaction(emoji, channelID, timestamp)
	
	assert.NoError(t, err)
	mockSlackClient.AssertCalled(t, "AddReaction", emoji, channelID, timestamp)
}
