package slack

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
	"go.uber.org/zap"
)

var _ EventProcessor = (*eventProcessorImpl)(nil)

type eventProcessorImpl struct {
	translationUseCase service.TranslationService
	slackClient        *SlackClient
	logger             *zap.Logger
}

func NewEventProcessor(
	translationUseCase service.TranslationService,
	slackClient *SlackClient,
	logger *zap.Logger,
) EventProcessor {
	return &eventProcessorImpl{
		translationUseCase: translationUseCase,
		slackClient:        slackClient,
		logger:             logger,
	}
}

func (ep *eventProcessorImpl) ProcessEvent(ctx context.Context, payload map[string]interface{}) {
	eventType, ok := payload["type"].(string)
	if !ok {
		ep.logger.Error("Failed to get event type")
		return
	}

	ep.logger.Info("Processing Slack event",
		zap.String("event_type", eventType))

	switch eventType {
	case "event_callback":
		ep.handleEventCallback(ctx, payload)
	default:
		ep.logger.Debug("Ignoring event type", zap.String("type", eventType))
	}
}

func (ep *eventProcessorImpl) handleEventCallback(ctx context.Context, payload map[string]interface{}) {
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

func (ep *eventProcessorImpl) handleMessageEvent(ctx context.Context, event map[string]interface{}) {
	// Skip messages with subtype (threaded replies, edits, etc.)
	if subtype, ok := event["subtype"].(string); ok && subtype != "" {
		ep.logger.Debug("Skipping message with subtype", zap.String("subtype", subtype))
		return
	}

	// Skip bot messages
	if _, ok := event["bot_id"].(string); ok {
		ep.logger.Debug("Skipping bot message")
		return
	}

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

	ts, tsOk := event["ts"].(string)
	if !tsOk {
		ep.logger.Error("Failed to get message timestamp")
		return
	}

	textPreview := text
	if len(text) > 50 {
		textPreview = text[:50]
	}

	ep.logger.Info("Processing message event",
		zap.String("channel_id", channelID),
		zap.String("user_id", userID),
		zap.String("text", textPreview),
		zap.String("timestamp", ts))

	// Add eye emoji reaction to the message
	if err := ep.slackClient.AddReaction("eyes", channelID, ts); err != nil {
		ep.logger.Warn("Failed to add emoji reaction to message",
			zap.Error(err),
			zap.String("channel_id", channelID),
			zap.String("timestamp", ts),
			zap.String("emoji", "eyes"),
			zap.String("troubleshooting", "Check if bot has reactions:write scope in Slack app OAuth settings"))
	}

	// Check if message contains only emoji codes
	if isEmojiOnly(text) {
		ep.logger.Info("Message contains only emoji, skipping translation",
			zap.String("text", text))
		return
	}

	// Detect message language using original text with emoji codes
	detectedLang, err := ep.detectLanguage(ctx, text)
	if err != nil {
		ep.logger.Error("Failed to detect message language",
			zap.Error(err),
			zap.String("text", text))
		return
	}

	ep.logger.Info("Language detected",
		zap.String("detected_language", detectedLang),
		zap.String("text", text[:min(len(text), 30)]))

	// Determine target language based on detected source language
	targetLang := "Vietnamese"
	if detectedLang == "Vietnamese" {
		targetLang = "English"
	} else if detectedLang != "English" {

		ep.logger.Info("Unsupported language, only English and Vietnamese are supported",
			zap.String("detected_language", detectedLang))

		// Post error message to thread
		errorMsg := "Sorry, I can't support this language. I only translate English and Vietnamese."
		_, _, err := ep.slackClient.PostMessage(channelID, errorMsg, ts)
		if err != nil {
			ep.logger.Error("Failed to post error message",
				zap.Error(err),
				zap.String("channel_id", channelID))
		}
		return
	}

	translationReq := request.Translation{
		Text:           text,
		SourceLanguage: detectedLang,
		TargetLanguage: targetLang,
	}

	result, err := ep.translationUseCase.Translate(translationReq)
	if err != nil {
		ep.logger.Error("Failed to translate message",
			zap.Error(err),
			zap.String("text", text))
		return
	}

	translatedText := result.TranslatedText

	ep.logger.Info("Translation completed",
		zap.String("original", text),
		zap.String("translated", translatedText),
		zap.String("source_lang", result.SourceLanguage),
		zap.String("target_lang", result.TargetLanguage))

	// Post translated message as a thread reply with emoji flag
	emoji := "🇻🇳"
	if result.TargetLanguage == "English" {
		emoji = "🇬🇧"
	}
	responseText := fmt.Sprintf("%s %s", emoji, translatedText)
	_, _, err = ep.slackClient.PostMessage(channelID, responseText, ts)
	if err != nil {
		ep.logger.Error("Failed to post translated message",
			zap.Error(err),
			zap.String("channel_id", channelID))
		return
	}

	ep.logger.Info("Translation posted successfully",
		zap.String("channel_id", channelID),
		zap.String("original", text[:min(len(text), 30)]),
		zap.String("translated", translatedText[:min(len(translatedText), 30)]))
}

func (ep *eventProcessorImpl) detectLanguage(ctx context.Context, text string) (string, error) {
	language, err := ep.translationUseCase.DetectLanguage(text)
	if err != nil {
		ep.logger.Error("Failed to detect language", zap.Error(err))
		return "", err
	}
	ep.logger.Debug("Language detection result",
		zap.String("detected_language", language))
	return language, nil
}

func (ep *eventProcessorImpl) handleReactionEvent(ctx context.Context, event map[string]interface{}) {
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

	// Add eye emoji reaction to the message
	if err := ep.slackClient.AddReaction("eyes", channelID, messageTS); err != nil {
		ep.logger.Warn("Failed to add emoji reaction to message",
			zap.Error(err),
			zap.String("channel_id", channelID),
			zap.String("timestamp", messageTS),
			zap.String("emoji", "eyes"),
			zap.String("troubleshooting", "Check if bot has reactions:write scope in Slack app OAuth settings"))
	}

	// Fetch the original message
	message, err := ep.slackClient.GetMessage(channelID, messageTS)
	if err != nil {
		ep.logger.Error("Failed to fetch message",
			zap.Error(err),
			zap.String("channel_id", channelID),
			zap.String("message_ts", messageTS))
		return
	}

	if message == nil || message.Text == "" {
		ep.logger.Warn("Message not found or empty",
			zap.String("channel_id", channelID),
			zap.String("message_ts", messageTS))
		return
	}

	// Check if message contains only emoji codes
	if isEmojiOnly(message.Text) {
		ep.logger.Info("Message contains only emoji, skipping translation from reaction",
			zap.String("text", message.Text))
		return
	}

	// Detect language from the message using original text with emoji codes
	detectedLang, err := ep.detectLanguage(ctx, message.Text)
	if err != nil {
		ep.logger.Error("Failed to detect message language",
			zap.Error(err),
			zap.String("text", message.Text))
		return
	}

	// Determine target language based on detected source language
	targetLang := "Vietnamese"
	if detectedLang == "Vietnamese" {
		targetLang = "English"
	} else if detectedLang != "English" {
		ep.logger.Info("Unsupported language, only English and Vietnamese are supported",
			zap.String("detected_language", detectedLang))

		// Post error message to thread
		errorMsg := "Sorry, I can't support this language. I only translate English and Vietnamese."
		_, _, err := ep.slackClient.PostMessage(channelID, errorMsg, messageTS)
		if err != nil {
			ep.logger.Error("Failed to post error message",
				zap.Error(err),
				zap.String("channel_id", channelID))
		}
		return
	}

	translationReq := request.Translation{
		Text:           message.Text,
		SourceLanguage: detectedLang,
		TargetLanguage: targetLang,
	}

	result, err := ep.translationUseCase.Translate(translationReq)
	if err != nil {
		ep.logger.Error("Failed to translate message from reaction",
			zap.Error(err),
			zap.String("text", message.Text))
		return
	}

	translatedText := result.TranslatedText

	// Post translated message as a thread reply with emoji flag
	emoji := "🇻🇳"
	if result.TargetLanguage == "English" {
		emoji = "🇬🇧"
	}
	responseText := fmt.Sprintf("%s %s", emoji, translatedText)
	_, _, err = ep.slackClient.PostMessage(channelID, responseText, messageTS)
	if err != nil {
		ep.logger.Error("Failed to post translated message from reaction",
			zap.Error(err),
			zap.String("channel_id", channelID))
		return
	}

	ep.logger.Info("Translation from reaction posted successfully",
		zap.String("channel_id", channelID),
		zap.String("original", message.Text[:min(len(message.Text), 30)]),
		zap.String("translated", translatedText[:min(len(translatedText), 30)]))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isEmojiOnly(text string) bool {
	emojiPattern := regexp.MustCompile(`:[a-zA-Z0-9_-]+:`)
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	// Remove all emoji codes
	withoutEmojis := emojiPattern.ReplaceAllString(trimmed, "")
	// Check if anything is left after removing emojis and whitespace
	return strings.TrimSpace(withoutEmojis) == ""
}

func extractEmojis(text string) (string, map[string]string) {
	emojiPattern := regexp.MustCompile(`:[a-zA-Z0-9_-]+:`)
	emojis := make(map[string]string)

	cleanedText := emojiPattern.ReplaceAllStringFunc(text, func(emoji string) string {
		placeholder := fmt.Sprintf("EMOJIPLACEHOLDER%d", len(emojis))
		emojis[placeholder] = emoji
		return placeholder
	})

	return cleanedText, emojis
}

func restoreEmojis(text string, emojis map[string]string) string {
	result := text
	for placeholder, emoji := range emojis {
		result = strings.ReplaceAll(result, placeholder, emoji)
	}
	return result
}
