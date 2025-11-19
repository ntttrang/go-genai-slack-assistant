package slack

import (
	"fmt"
	"strings"

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
	if sc.client == nil {
		return nil, fmt.Errorf("slack client is not initialized")
	}

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
	return sc.PostMessageWithBotInfo(channelID, text, threadTS, "", "")
}

func (sc *SlackClient) PostMessageWithBotInfo(channelID, text string, threadTS string, username string, avatarURL string) (string, string, error) {
	if sc.client == nil {
		return "", "", fmt.Errorf("slack client is not initialized")
	}

	opts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
	}

	if threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(threadTS))
	}

	if username != "" {
		opts = append(opts, slack.MsgOptionUsername(username))
	}

	if avatarURL != "" {
		opts = append(opts, slack.MsgOptionIconURL(avatarURL))
	}

	channel, ts, err := sc.client.PostMessage(channelID, opts...)
	return channel, ts, err
}

func (sc *SlackClient) PostMessageWithBotInfoAndFiles(channelID, text string, threadTS string, username string, avatarURL string, files []FileInfo) (string, string, error) {
	if sc.client == nil {
		return "", "", fmt.Errorf("slack client is not initialized")
	}

	opts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
	}

	if threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(threadTS))
	}

	if username != "" {
		opts = append(opts, slack.MsgOptionUsername(username))
	}

	if avatarURL != "" {
		opts = append(opts, slack.MsgOptionIconURL(avatarURL))
	}

	// Add blocks for images/files if present
	if len(files) > 0 {
		blocks := []slack.Block{}

		// Add text block first
		textBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", text, false, false),
			nil,
			nil,
		)
		blocks = append(blocks, textBlock)

		// Add context blocks for all files (images and documents)
		// Note: We use permalinks instead of url_private because Image Blocks
		// cannot access Slack's private URLs (they require auth headers)
		for _, file := range files {
			if file.Permalink != "" {
				// Use different emoji for images vs other files
				emoji := "📎"
				if strings.HasPrefix(file.Mimetype, "image/") {
					emoji = "🖼️"
				}
				contextText := fmt.Sprintf("%s <%s|%s>", emoji, file.Permalink, file.Name)
				contextBlock := slack.NewContextBlock("",
					slack.NewTextBlockObject("mrkdwn", contextText, false, false),
				)
				blocks = append(blocks, contextBlock)
			}
		}

		// Replace text option with blocks
		opts = []slack.MsgOption{
			slack.MsgOptionBlocks(blocks...),
		}

		if threadTS != "" {
			opts = append(opts, slack.MsgOptionTS(threadTS))
		}

		if username != "" {
			opts = append(opts, slack.MsgOptionUsername(username))
		}

		if avatarURL != "" {
			opts = append(opts, slack.MsgOptionIconURL(avatarURL))
		}
	}

	channel, ts, err := sc.client.PostMessage(channelID, opts...)
	return channel, ts, err
}

// PostMessageWithBotInfoAsQuote posts a message as a quote (with left border) using blocks
func (sc *SlackClient) PostMessageWithBotInfoAsQuote(channelID, text string, threadTS string, username string, avatarURL string) (string, string, error) {
	if sc.client == nil {
		return "", "", fmt.Errorf("slack client is not initialized")
	}

	opts := []slack.MsgOption{
		slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "> "+text, false, false),
				nil,
				nil,
			),
		),
	}

	if threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(threadTS))
	}

	if username != "" {
		opts = append(opts, slack.MsgOptionUsername(username))
	}

	if avatarURL != "" {
		opts = append(opts, slack.MsgOptionIconURL(avatarURL))
	}

	channel, ts, err := sc.client.PostMessage(channelID, opts...)
	return channel, ts, err
}

// PostMessageWithBotInfoAsQuoteAndFiles posts a quote message with files
func (sc *SlackClient) PostMessageWithBotInfoAsQuoteAndFiles(channelID, text string, threadTS string, username string, avatarURL string, files []FileInfo) (string, string, error) {
	if sc.client == nil {
		return "", "", fmt.Errorf("slack client is not initialized")
	}

	blocks := []slack.Block{}

	// Add text block as quote
	textBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", "> "+text, false, false),
		nil,
		nil,
	)
	blocks = append(blocks, textBlock)

	// Add context blocks for all files (images and documents)
	for _, file := range files {
		if file.Permalink != "" {
			// Use different emoji for images vs other files
			emoji := "📎"
			if strings.HasPrefix(file.Mimetype, "image/") {
				emoji = "🖼️"
			}
			contextText := fmt.Sprintf("%s <%s|%s>", emoji, file.Permalink, file.Name)
			contextBlock := slack.NewContextBlock("",
				slack.NewTextBlockObject("mrkdwn", contextText, false, false),
			)
			blocks = append(blocks, contextBlock)
		}
	}

	opts := []slack.MsgOption{
		slack.MsgOptionBlocks(blocks...),
	}

	if threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(threadTS))
	}

	if username != "" {
		opts = append(opts, slack.MsgOptionUsername(username))
	}

	if avatarURL != "" {
		opts = append(opts, slack.MsgOptionIconURL(avatarURL))
	}

	channel, ts, err := sc.client.PostMessage(channelID, opts...)
	return channel, ts, err
}

func (sc *SlackClient) GetUserInfo(userID string) (*slack.User, error) {
	if sc.client == nil {
		return nil, fmt.Errorf("slack client is not initialized")
	}
	return sc.client.GetUserInfo(userID)
}

func (sc *SlackClient) AddReaction(emoji, channelID, timestamp string) error {
	if sc.client == nil {
		return nil // Silently return nil in test scenarios
	}
	return sc.client.AddReaction(emoji, slack.ItemRef{
		Channel:   channelID,
		Timestamp: timestamp,
	})
}
