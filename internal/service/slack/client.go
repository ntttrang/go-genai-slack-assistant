package slack

import (
	"github.com/slack-go/slack"
)

type SlackClient struct {
	client *slack.Client
}

func NewSlackClient(token string) *SlackClient {
	return &SlackClient{
		client: slack.New(token),
	}
}

func (sc *SlackClient) GetMessage(channelID, timestamp string) (*slack.Message, error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Latest:    timestamp,
		Inclusive: true,
		Limit:     1,
	}

	history, err := sc.client.GetConversationHistory(params)
	if err != nil {
		return nil, err
	}

	if len(history.Messages) == 0 {
		return nil, nil
	}

	// Verify we got the right message by timestamp
	if history.Messages[0].Timestamp != timestamp {
		return nil, nil
	}

	return &history.Messages[0], nil
}

func (sc *SlackClient) PostMessage(channelID, text string, threadTS string) (string, string, error) {
	opts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
	}

	if threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(threadTS))
	}

	channel, ts, err := sc.client.PostMessage(channelID, opts...)
	return channel, ts, err
}

func (sc *SlackClient) GetUserInfo(userID string) (*slack.User, error) {
	return sc.client.GetUserInfo(userID)
}

func (sc *SlackClient) AddReaction(emoji, channelID, timestamp string) error {
	return sc.client.AddReaction(emoji, slack.ItemRef{
		Channel:   channelID,
		Timestamp: timestamp,
	})
}
