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
	default:
		ep.logger.Debug("Ignoring callback event type", zap.String("type", eventType))
	}
}

func (ep *eventProcessorImpl) handleMessageEvent(ctx context.Context, event map[string]interface{}) {
	// Skip messages with certain subtypes (threaded replies, edits, etc.)
	// But allow file_share subtype (messages with images/files)
	if subtype, ok := event["subtype"].(string); ok && subtype != "" {
		// Allow file_share subtype to be processed
		if subtype != "file_share" {
			ep.logger.Debug("Skipping message with subtype", zap.String("subtype", subtype))
			return
		}
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

	text, ok := event["text"].(string)
	if !ok {
		text = ""
	}

	// Trim whitespace to check if there's actual text content
	trimmedText := strings.TrimSpace(text)

	// Check if message contains files
	hasFiles := false
	if filesInterface, ok := event["files"]; ok {
		if filesArray, ok := filesInterface.([]interface{}); ok && len(filesArray) > 0 {
			hasFiles = true
		}
	}

	// If message has files but no text, just add eyes reaction and return
	if hasFiles && trimmedText == "" {
		ep.logger.Info("Message contains files only (no text), adding eyes reaction",
			zap.String("channel_id", channelID),
			zap.String("user_id", userID),
			zap.String("timestamp", ts))

		if err := ep.slackClient.AddReaction("eyes", channelID, ts); err != nil {
			ep.logger.Warn("Failed to add emoji reaction to message",
				zap.Error(err),
				zap.String("channel_id", channelID),
				zap.String("timestamp", ts),
				zap.String("emoji", "eyes"),
				zap.String("troubleshooting", "Check if bot has reactions:write scope in Slack app OAuth settings"))
		}
		return
	}

	// If no text at all (and no files), skip
	if trimmedText == "" {
		ep.logger.Debug("Skipping message with empty or missing text",
			zap.String("channel_id", channelID),
			zap.Any("event", event))
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

	// Check if message contains only user mentions or @here/@channel
	if isUserMentionOnly(text) {
		ep.logger.Info("Message contains only mentions (@user, @here, @channel), skipping translation",
			zap.String("text", text))
		return
	}

	// Get user info for custom bot name and avatar
	userInfo, err := ep.slackClient.GetUserInfo(userID)
	botName := "SlackBot"
	botAvatar := ""
	if err == nil && userInfo != nil {
		displayName := userInfo.Profile.DisplayName
		if displayName == "" {
			displayName = userInfo.Name
		}
		// botName = fmt.Sprintf("%s (Bot) %s", displayName, emoji)
		botName = fmt.Sprintf("%s (Bot)", displayName)
		botAvatar = userInfo.Profile.Image512
		if botAvatar == "" {
			botAvatar = userInfo.Profile.Image48
		}
		ep.logger.Debug("User info retrieved",
			zap.String("user_name", userInfo.Name),
			zap.String("bot_name", botName))
	} else {
		ep.logger.Warn("Failed to get user info, using default bot name",
			zap.Error(err))
	}

	// Detect message language using original text with emoji codes
	detectedLang, err := ep.detectLanguage(ctx, text)
	if err != nil {
		ep.logger.Error("Failed to detect message language",
			zap.Error(err),
			zap.String("text", text))

		// Check if quota exceeded error
		if strings.Contains(err.Error(), "googleapi: Error 429: Resource exhausted") {
			errorMessage := "❌ Sorry, I can't translate because the current quota has been exceeded. Please try again later."
			_, _, err = ep.slackClient.PostMessageWithBotInfo(channelID, errorMessage, ts, botName, botAvatar)
			if err != nil {
				ep.logger.Error("Failed to post error message",
					zap.Error(err),
					zap.String("channel_id", channelID))
			}
			return
		}
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
		errorMsg := "⚠️ Sorry! I only translate English and Vietnamese right now, not other languages, slang or numbers"
		_, _, err = ep.slackClient.PostMessageWithBotInfo(channelID, errorMsg, ts, botName, botAvatar)
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
		if strings.Contains(err.Error(), "Delimiter tag injection") || strings.Contains(err.Error(), "input validation failed") {
			ep.logger.Warn("Security validation failed for message",
				zap.Error(err),
				zap.String("channel_id", channelID),
				zap.String("user_id", userID))

			errorMsg := "Sorry, there seems to be an error in your text. Please check the content and try again."
			_, _, postErr := ep.slackClient.PostMessageWithBotInfo(channelID, errorMsg, ts, botName, botAvatar)
			if postErr != nil {
				ep.logger.Error("Failed to post security error message",
					zap.Error(postErr),
					zap.String("channel_id", channelID))
			}
			return
		}

		ep.logger.Error("Failed to translate message",
			zap.Error(err),
			zap.String("text", text))

		errorMsg := "❌ Sorry, I can't translate because the current quota has been exceeded. Please try again later."
		_, _, postErr := ep.slackClient.PostMessageWithBotInfo(channelID, errorMsg, ts, botName, botAvatar)
		if postErr != nil {
			ep.logger.Error("Failed to post translation error message",
				zap.Error(postErr),
				zap.String("channel_id", channelID))
		}
		return
	}

	translatedText := result.TranslatedText

	// Convert user mentions and @here/@channel to quoted format
	translatedText = strings.ReplaceAll(translatedText, "<!here>", "`here`")
	translatedText = strings.ReplaceAll(translatedText, "<!channel>", "`channel`")
	// // Quote user mentions like <@USERID|username>
	userMentionPattern := regexp.MustCompile(`<@[^>]+>`)
	translatedText = userMentionPattern.ReplaceAllStringFunc(translatedText, func(match string) string {
		return "`" + match + "`"
	})

	responseText := translatedText

	// Customize botName
	//Determine emoji flag based on target language
	emoji := "🇻🇳"
	if result.TargetLanguage == "English" {
		emoji = "🇬🇧"
	}
	botName = fmt.Sprintf("%s %s", botName, emoji)

	// Extract files from the original message event
	files := ep.extractFiles(event)

	// Check if message contains @here or @channel tags
	isQuote := containsAtHereOrChannel(text)

	// Post message with appropriate format (quote or normal)
	if isQuote {
		if len(files) > 0 {
			_, _, err = ep.slackClient.PostMessageWithBotInfoAsQuoteAndFiles(channelID, responseText, ts, botName, botAvatar, files)
		} else {
			_, _, err = ep.slackClient.PostMessageWithBotInfoAsQuote(channelID, responseText, ts, botName, botAvatar)
		}
	} else {
		_, _, err = ep.slackClient.PostMessageWithBotInfoAndFiles(channelID, responseText, ts, botName, botAvatar, files)
	}

	if err != nil {
		ep.logger.Error("Failed to post translated message",
			zap.Error(err),
			zap.String("channel_id", channelID))
		return
	}

	ep.logger.Info("Translation posted successfully",
		zap.String("channel_id", channelID),
		zap.String("original", text[:min(len(text), 30)]),
		zap.String("translated", translatedText[:min(len(translatedText), 30)]),
		zap.Bool("is_quote", isQuote))
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

func isUserMentionOnly(text string) bool {
	userMentionPattern := regexp.MustCompile(`<@[^>]+>`)
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	// Remove all user mentions
	withoutUserMentions := userMentionPattern.ReplaceAllString(trimmed, "")
	// Remove @here and @channel mentions (both Slack tags and @ format)
	withoutHereChannel := strings.ReplaceAll(withoutUserMentions, "<!here>", "")
	withoutHereChannel = strings.ReplaceAll(withoutHereChannel, "<!channel>", "")
	withoutHereChannel = strings.ReplaceAll(withoutHereChannel, "@here", "")
	withoutHereChannel = strings.ReplaceAll(withoutHereChannel, "@channel", "")
	// Check if anything is left after removing all mentions and whitespace
	return strings.TrimSpace(withoutHereChannel) == ""
}

// containsAtHereOrChannel checks if message contains @here or @channel tags
func containsAtHereOrChannel(text string) bool {
	return strings.Contains(text, "<!here>") || strings.Contains(text, "<!channel>") ||
		strings.Contains(text, "@here") || strings.Contains(text, "@channel")
}

// extractMentionPrefix extracts @here or @channel mention from the beginning of text
// Returns the mention in backtick format (e.g., "`here`" or "`channel`") or empty string if no mention found
func (ep *eventProcessorImpl) extractMentionPrefix(text string) string {
	trimmed := strings.TrimSpace(text)

	// Check for <!here> or <!channel> tags
	if strings.HasPrefix(trimmed, "<!here>") {
		return "`here`"
	}
	if strings.HasPrefix(trimmed, "<!channel>") {
		return "`channel`"
	}

	// Check for @here or @channel
	if strings.HasPrefix(trimmed, "@here") {
		return "`here`"
	}
	if strings.HasPrefix(trimmed, "@channel") {
		return "`channel`"
	}

	return ""
}

// convertUserMentionsToText converts mentions in the original text to backticks in the translated text
// If original text has @here/@channel at the start, it prepends them to the translated text
func (ep *eventProcessorImpl) convertUserMentionsToText(originalText, translatedText string) string {
	// Extract mention prefix from original text
	mentionPrefix := ep.extractMentionPrefix(originalText)

	// If we found a mention prefix, prepend it to the translated text
	if mentionPrefix != "" {
		return mentionPrefix + " " + translatedText
	}

	// Return translated text as-is
	return translatedText
}

// convertHereChannelMentionsToQuotes converts @here/@channel and <!here>/<!channel> mentions to backtick format
func (ep *eventProcessorImpl) convertHereChannelMentionsToQuotes(text string) string {
	result := text

	// Convert Slack tags
	result = strings.ReplaceAll(result, "<!here>", "`here`")
	result = strings.ReplaceAll(result, "<!channel>", "`channel`")

	// Convert @ mentions
	result = strings.ReplaceAll(result, "@here", "`here`")
	result = strings.ReplaceAll(result, "@channel", "`channel`")

	return result
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

// FileInfo represents file information from Slack
type FileInfo struct {
	URL       string
	Permalink string
	Mimetype  string
	Name      string
}

// extractFiles extracts file information from a Slack event
func (ep *eventProcessorImpl) extractFiles(event map[string]interface{}) []FileInfo {
	files := []FileInfo{}

	filesInterface, ok := event["files"]
	if !ok {
		return files
	}

	filesArray, ok := filesInterface.([]interface{})
	if !ok {
		return files
	}

	for _, fileInterface := range filesArray {
		fileMap, ok := fileInterface.(map[string]interface{})
		if !ok {
			continue
		}

		fileInfo := FileInfo{}

		// Extract URL and permalink - both are useful
		if urlPrivate, ok := fileMap["url_private"].(string); ok {
			fileInfo.URL = urlPrivate
		}

		if permalink, ok := fileMap["permalink"].(string); ok {
			fileInfo.Permalink = permalink
		}

		if mimetype, ok := fileMap["mimetype"].(string); ok {
			fileInfo.Mimetype = mimetype
		}

		if name, ok := fileMap["name"].(string); ok {
			fileInfo.Name = name
		}

		// Only add if we have at least a URL or permalink
		if fileInfo.URL != "" || fileInfo.Permalink != "" {
			files = append(files, fileInfo)
			ep.logger.Debug("Extracted file from event",
				zap.String("name", fileInfo.Name),
				zap.String("mimetype", fileInfo.Mimetype))
		}
	}

	return files
}
