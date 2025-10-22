package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"github.com/slack-translation-bot/domain"
	"github.com/slack-translation-bot/usecase"
)

type EventProcessor struct {
	translationUseCase *usecase.TranslationUseCase
	slackClient        *SlackClient
	logger             *zap.Logger
}

func NewEventProcessor(
	translationUseCase *usecase.TranslationUseCase,
	slackClient *SlackClient,
	logger *zap.Logger,
) *EventProcessor {
	return &EventProcessor{
		translationUseCase: translationUseCase,
		slackClient:        slackClient,
		logger:             logger,
	}
}

func (ep *EventProcessor) ProcessEvent(ctx context.Context, payload map[string]interface{}) {
	eventType, ok := payload["type"].(string)
	if !ok {
		ep.logger.Error("Failed to get event type")
		return
	}

	switch eventType {
	case "event_callback":
		ep.handleEventCallback(ctx, payload)
	default:
		ep.logger.Debug("Ignoring event type", zap.String("type", eventType))
	}
}

func (ep *EventProcessor) handleEventCallback(ctx context.Context, payload map[string]interface{}) {
	event, ok := payload["event"].(map[string]interface{})
	if !ok {
		ep.logger.Error("Failed to get event data")
		return
	}

	eventType, ok := event["type"].(string)
	if !ok {
		ep.logger.Error("Failed to get event type from callback")
		return
	}

	switch eventType {
	case "message":
		ep.handleMessageEvent(ctx, event)
	case "reaction_added":
		ep.handleReactionEvent(ctx, event)
	default:
		ep.logger.Debug("Ignoring callback event type", zap.String("type", eventType))
	}
}

func (ep *EventProcessor) handleMessageEvent(ctx context.Context, event map[string]interface{}) {
	channelID, ok := event["channel"].(string)
	if !ok {
		ep.logger.Error("Failed to get channel ID")
		return
	}

	text, ok := event["text"].(string)
	if !ok || text == "" {
		return
	}

	userID, ok := event["user"].(string)
	if !ok {
		ep.logger.Error("Failed to get user ID")
		return
	}

	timestamp, ok := event["ts"].(string)
	if !ok {
		ep.logger.Error("Failed to get message timestamp")
		return
	}

	ep.logger.Info("Processing message event",
		zap.String("channel_id", channelID),
		zap.String("user_id", userID),
		zap.String("text", text[:min(len(text), 50)]))

	// Store message context for later processing
	// TODO: Implement auto-translation based on channel config
}

func (ep *EventProcessor) handleReactionEvent(ctx context.Context, event map[string]interface{}) {
	reaction, ok := event["reaction"].(string)
	if !ok {
		ep.logger.Error("Failed to get reaction")
		return
	}

	// Check if reaction is Vietnamese flag emoji
	if reaction != "flag-vn" && reaction != "vn" {
		return
	}

	itemType, ok := event["item"].(map[string]interface{})
	if !ok {
		ep.logger.Error("Failed to get item data")
		return
	}

	channelID, ok := itemType["channel"].(string)
	if !ok {
		ep.logger.Error("Failed to get channel ID from reaction")
		return
	}

	messageTS, ok := itemType["ts"].(string)
	if !ok {
		ep.logger.Error("Failed to get message timestamp from reaction")
		return
	}

	ep.logger.Info("Processing reaction event",
		zap.String("channel_id", channelID),
		zap.String("reaction", reaction),
		zap.String("message_ts", messageTS))

	// TODO: Trigger translation for the message
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
