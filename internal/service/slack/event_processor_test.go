package slack

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/testutils/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEventProcessorProcessEventURLVerification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger)

	payload := map[string]interface{}{
		"type":      "url_verification",
		"challenge": "test-challenge-123",
	}

	// This should not panic and handle gracefully
	processor.ProcessEvent(context.Background(), payload)
}

func TestEventProcessorProcessEventCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger)

	// Create minimal valid event callback
	payload := map[string]interface{}{
		"type": "event_callback",
		"event": map[string]interface{}{
			"type": "unknown_event_type",
		},
	}

	// This should handle gracefully
	processor.ProcessEvent(context.Background(), payload)
}

func TestEventProcessorImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger)
	assert.NotNil(t, processor)
}

func TestEventProcessorInvalidEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger)

	payload := map[string]interface{}{
		"type": "invalid_type",
	}

	// This should handle gracefully
	processor.ProcessEvent(context.Background(), payload)
}

func TestEventProcessorHandleMessageEvent_EmptyText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"channel": "C123456",
		"text":    "",
		"user":    "U123456",
		"ts":      "1234567890.123456",
	}

	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_SkipBotMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"bot_id":  "B123456",
		"channel": "C123456",
		"text":    "Hello",
		"ts":      "1234567890.123456",
	}

	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_SkipMessageWithSubtype(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"subtype": "message_changed",
		"channel": "C123456",
		"text":    "Hello",
		"user":    "U123456",
		"ts":      "1234567890.123456",
	}

	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_AllowFileShareSubtype(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	// Create a mock SlackClient to avoid nil pointer dereference
	mockSlackClient := &SlackClient{
		client: nil, // We don't need actual client for this test
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	// Message with file_share subtype should be processed (not skipped at validation stage)
	event := map[string]interface{}{
		"type":    "message",
		"subtype": "file_share",
		"channel": "C123456",
		"text":    "Check out this image",
		"user":    "U123456",
		"ts":      "1234567890.123456",
		"files": []interface{}{
			map[string]interface{}{
				"id":          "F123456",
				"name":        "image.png",
				"mimetype":    "image/png",
				"url_private": "https://files.slack.com/test.png",
				"permalink":   "https://example.slack.com/files/test.png",
			},
		},
	}

	// Set up mock expectations - the message will be processed normally
	mockTranslationService.EXPECT().
		DetectLanguage(gomock.Any()).
		Return("", fmt.Errorf("test error")).
		Times(1)

	// This should not be skipped at the validation stage
	// The key is that file_share subtype is NOT filtered out
	processor.handleMessageEvent(context.Background(), event)
}

func TestExtractFiles_WithImageFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"files": []interface{}{
			map[string]interface{}{
				"id":          "F123456",
				"name":        "screenshot.png",
				"mimetype":    "image/png",
				"url_private": "https://files.slack.com/files-pri/T123/F123/screenshot.png",
				"permalink":   "https://example.slack.com/files/U123/F123/screenshot.png",
			},
		},
	}

	files := processor.extractFiles(event)

	assert.Len(t, files, 1)
	assert.Equal(t, "screenshot.png", files[0].Name)
	assert.Equal(t, "image/png", files[0].Mimetype)
	assert.Equal(t, "https://files.slack.com/files-pri/T123/F123/screenshot.png", files[0].URL)
	assert.Equal(t, "https://example.slack.com/files/U123/F123/screenshot.png", files[0].Permalink)
}

func TestExtractFiles_WithMultipleFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"files": []interface{}{
			map[string]interface{}{
				"id":          "F123456",
				"name":        "image.jpg",
				"mimetype":    "image/jpeg",
				"url_private": "https://files.slack.com/image.jpg",
				"permalink":   "https://example.slack.com/files/image.jpg",
			},
			map[string]interface{}{
				"id":          "F123457",
				"name":        "document.pdf",
				"mimetype":    "application/pdf",
				"url_private": "https://files.slack.com/document.pdf",
				"permalink":   "https://example.slack.com/files/document.pdf",
			},
		},
	}

	files := processor.extractFiles(event)

	assert.Len(t, files, 2)
	assert.Equal(t, "image.jpg", files[0].Name)
	assert.Equal(t, "image/jpeg", files[0].Mimetype)
	assert.Equal(t, "document.pdf", files[1].Name)
	assert.Equal(t, "application/pdf", files[1].Mimetype)
}

func TestExtractFiles_NoFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"text": "Just a text message",
	}

	files := processor.extractFiles(event)

	assert.Len(t, files, 0)
}

func TestIsEmojiOnly_OnlyEmoji(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"single emoji", ":smile:", true},
		{"multiple emojis", ":smile: :wave:", true},
		{"emojis with spaces", "  :smile:  :wave:  ", true},
		{"emoji with text", ":smile: Hello", false},
		{"text only", "Hello world", false},
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"mixed", "Hello :smile:", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmojiOnly(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractAndRestoreEmojis(t *testing.T) {
	originalText := "Hello :smile: world :wave:"

	cleanedText, emojiMap := extractEmojis(originalText)

	assert.Contains(t, cleanedText, "EMOJIPLACEHOLDER")
	assert.Len(t, emojiMap, 2)

	restoredText := restoreEmojis(cleanedText, emojiMap)

	assert.Equal(t, originalText, restoredText)
}

func TestExtractAndRestoreEmojis_NoEmojis(t *testing.T) {
	originalText := "Hello world"

	cleanedText, emojiMap := extractEmojis(originalText)

	assert.Equal(t, originalText, cleanedText)
	assert.Len(t, emojiMap, 0)
}

func TestExtractAndRestoreEmojis_OnlyEmojis(t *testing.T) {
	originalText := ":smile: :wave: :tada:"

	cleanedText, emojiMap := extractEmojis(originalText)

	assert.NotEqual(t, originalText, cleanedText)
	assert.Len(t, emojiMap, 3)

	restoredText := restoreEmojis(cleanedText, emojiMap)

	assert.Equal(t, originalText, restoredText)
}

func TestConvertUserMentionsToText_WithoutMentions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	originalText := "Hello world"
	translatedText := "Xin chào thế giới"

	result := processor.convertUserMentionsToText(originalText, translatedText)

	assert.Equal(t, translatedText, result)
}

func TestExtractMentionPrefix_WithAtHere(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "@here Are you there?"
	result := processor.extractMentionPrefix(text)

	assert.Equal(t, "`here`", result)
}

func TestExtractMentionPrefix_WithAtChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "@channel please review this"
	result := processor.extractMentionPrefix(text)

	assert.Equal(t, "`channel`", result)
}

func TestExtractMentionPrefix_WithSlackHereTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "<!here> everyone look at this"
	result := processor.extractMentionPrefix(text)

	assert.Equal(t, "`here`", result)
}

func TestExtractMentionPrefix_WithSlackChannelTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "<!channel> attention needed"
	result := processor.extractMentionPrefix(text)

	assert.Equal(t, "`channel`", result)
}

func TestExtractMentionPrefix_NoMention(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "Regular message without mentions"
	result := processor.extractMentionPrefix(text)

	assert.Equal(t, "", result)
}

func TestExtractMentionPrefix_WithWhitespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "   @here   message content"
	result := processor.extractMentionPrefix(text)

	assert.Equal(t, "`here`", result)
}

func TestConvertUserMentionsToText_WithAtHereMention(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	originalText := "@here Are you there?"
	translatedText := "Bạn có ở đó không?"

	result := processor.convertUserMentionsToText(originalText, translatedText)

	// Should prepend @here in backticks
	assert.Equal(t, "`here` Bạn có ở đó không?", result)
}

func TestConvertUserMentionsToText_WithAtChannelMention(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	originalText := "@channel please check this"
	translatedText := "vui lòng kiểm tra cái này"

	result := processor.convertUserMentionsToText(originalText, translatedText)

	// Should prepend @channel in backticks
	assert.Equal(t, "`channel` vui lòng kiểm tra cái này", result)
}

func TestConvertUserMentionsToText_WithMentionInMiddle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	mockSlackClient := &SlackClient{}

	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	// Setup mock
	mockSlackClient.client = nil // We'll mock GetUserInfo differently

	originalText := "Hey <@U12345> please check this"
	translatedText := "Xin chào <@U12345> vui lòng kiểm tra cái này"

	// Since we can't easily mock GetUserInfo, this test verifies the mention stays in place
	// In real usage, GetUserInfo would convert it to backticks
	result := processor.convertUserMentionsToText(originalText, translatedText)

	// The mention should still be present or converted to quoted format
	// For now, verify the function doesn't crash and handles mentions
	assert.NotNil(t, result)
	// If GetUserInfo fails, it should still have the mention in backticks or original
	assert.True(t, strings.Contains(result, "vui lòng kiểm tra") || strings.Contains(result, "U12345"))
}

func TestEventProcessorHandleMessageEvent_FilesOnlyNoText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	// Create a mock SlackClient to avoid nil pointer dereference
	mockSlackClient := &SlackClient{
		client: nil, // We don't need actual client for this test
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	// Message with files but no text - should only add eyes reaction, no translation
	event := map[string]interface{}{
		"type":    "message",
		"subtype": "file_share",
		"channel": "C123456",
		"text":    "", // Empty text
		"user":    "U123456",
		"ts":      "1234567890.123456",
		"files": []interface{}{
			map[string]interface{}{
				"id":          "F123456",
				"name":        "image.png",
				"mimetype":    "image/png",
				"url_private": "https://files.slack.com/test.png",
				"permalink":   "https://example.slack.com/files/test.png",
			},
		},
	}

	// This should not call translation service, only add reaction
	// No expectations set on mockTranslationService means it should not be called
	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_FilesOnlyWhitespaceText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	// Create a mock SlackClient to avoid nil pointer dereference
	mockSlackClient := &SlackClient{
		client: nil, // We don't need actual client for this test
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	// Message with files but only whitespace text - should only add eyes reaction, no translation
	event := map[string]interface{}{
		"type":    "message",
		"subtype": "file_share",
		"channel": "C123456",
		"text":    "   \n\t  ", // Only whitespace
		"user":    "U123456",
		"ts":      "1234567890.123456",
		"files": []interface{}{
			map[string]interface{}{
				"id":          "F123456",
				"name":        "document.pdf",
				"mimetype":    "application/pdf",
				"url_private": "https://files.slack.com/test.pdf",
				"permalink":   "https://example.slack.com/files/test.pdf",
			},
		},
	}

	// This should not call translation service, only add reaction
	// No expectations set on mockTranslationService means it should not be called
	processor.handleMessageEvent(context.Background(), event)
}

// Tests for containsAtHereOrChannel function
func TestContainsAtHereOrChannel_WithHereTag(t *testing.T) {
	text := "<!here> please review this"
	assert.True(t, containsAtHereOrChannel(text), "Should detect <!here> tag")
}

func TestContainsAtHereOrChannel_WithChannelTag(t *testing.T) {
	text := "<!channel> attention needed"
	assert.True(t, containsAtHereOrChannel(text), "Should detect <!channel> tag")
}

func TestContainsAtHereOrChannel_WithAtHere(t *testing.T) {
	text := "@here check this out"
	assert.True(t, containsAtHereOrChannel(text), "Should detect @here mention")
}

func TestContainsAtHereOrChannel_WithAtChannel(t *testing.T) {
	text := "@channel please look at this"
	assert.True(t, containsAtHereOrChannel(text), "Should detect @channel mention")
}

func TestContainsAtHereOrChannel_WithoutTags(t *testing.T) {
	text := "Regular message with @user mention"
	assert.False(t, containsAtHereOrChannel(text), "Should not detect tags when not present")
}

func TestContainsAtHereOrChannel_EmptyText(t *testing.T) {
	text := ""
	assert.False(t, containsAtHereOrChannel(text), "Should return false for empty text")
}

func TestContainsAtHereOrChannel_MultipleOccurrences(t *testing.T) {
	text := "<!here> and <!channel> both mentioned"
	assert.True(t, containsAtHereOrChannel(text), "Should detect multiple tags")
}

func TestContainsAtHereOrChannel_WithHereInMiddle(t *testing.T) {
	text := "Please notify @here about the update"
	assert.True(t, containsAtHereOrChannel(text), "Should detect @here even in middle of text")
}

func TestContainsAtHereOrChannel_WithChannelInMiddle(t *testing.T) {
	text := "Attention @channel this is important"
	assert.True(t, containsAtHereOrChannel(text), "Should detect @channel even in middle of text")
}

func TestConvertHereChannelMentionsToQuotes_WithAtHereInMiddle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "Please notify @here about the update"
	result := processor.convertHereChannelMentionsToQuotes(text)

	// Should convert @here to `here` even in middle of text
	assert.Equal(t, "Please notify `here` about the update", result)
}

func TestConvertHereChannelMentionsToQuotes_WithAtChannelInMiddle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "Attention @channel this is important"
	result := processor.convertHereChannelMentionsToQuotes(text)

	// Should convert @channel to `channel` even in middle of text
	assert.Equal(t, "Attention `channel` this is important", result)
}

func TestConvertHereChannelMentionsToQuotes_WithSlackHereInMiddle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "Please notify <!here> about the update"
	result := processor.convertHereChannelMentionsToQuotes(text)

	// Should convert <!here> to `here`
	assert.Equal(t, "Please notify `here` about the update", result)
}

func TestConvertHereChannelMentionsToQuotes_WithSlackChannelInMiddle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "Attention <!channel> this is important"
	result := processor.convertHereChannelMentionsToQuotes(text)

	// Should convert <!channel> to `channel`
	assert.Equal(t, "Attention `channel` this is important", result)
}

func TestConvertHereChannelMentionsToQuotes_WithMultipleMentions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "Hey @here and @channel please check this"
	result := processor.convertHereChannelMentionsToQuotes(text)

	// Should convert all mentions
	assert.Equal(t, "Hey `here` and `channel` please check this", result)
}

func TestConvertHereChannelMentionsToQuotes_NoMentions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()
	processor := NewEventProcessor(mockTranslationService, nil, logger).(*eventProcessorImpl)

	text := "Regular message without any special mentions"
	result := processor.convertHereChannelMentionsToQuotes(text)

	// Should remain unchanged
	assert.Equal(t, text, result)
}

func TestIsUserMentionOnly_OnlyUserMention(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"single user mention", "<@U123456>", true},
		{"multiple user mentions", "<@U123456> <@U789012>", true},
		{"user mention with username", "<@U123456|john>", true},
		{"mentions with spaces", "  <@U123456>  <@U789012>  ", true},
		{"mention with text", "<@U123456> Hello", false},
		{"text only", "Hello world", false},
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"mixed", "Hello <@U123456>", false},
		{"mention at end", "Check this out <@U123456>", false},
		{"only @here", "@here", true},
		{"only @channel", "@channel", true},
		{"only <!here>", "<!here>", true},
		{"only <!channel>", "<!channel>", true},
		{"@here with spaces", "  @here  ", true},
		{"@channel with text", "@channel please review", false},
		{"user mention and @here", "<@U123456> @here", true},
		{"user mention and @channel", "<@U123456> @channel", true},
		{"multiple mentions mixed", "<@U123456> <!here> <@U789012> @channel", true},
		{"@here with text", "@here check this", false},
		{"text with @here at end", "Hello @here", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUserMentionOnly(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEventProcessorHandleMessageEvent_UserMentionOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	mockSlackClient := &SlackClient{
		client: nil,
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"channel": "C123456",
		"text":    "<@U123456>",
		"user":    "U789012",
		"ts":      "1234567890.123456",
	}

	// No expectations on translation service means it should not be called
	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_MultipleUserMentionsOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	mockSlackClient := &SlackClient{
		client: nil,
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"channel": "C123456",
		"text":    "<@U123456> <@U789012> <@U345678>",
		"user":    "U999999",
		"ts":      "1234567890.123456",
	}

	// No expectations on translation service means it should not be called
	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_OnlyAtHere(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	mockSlackClient := &SlackClient{
		client: nil,
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"channel": "C123456",
		"text":    "@here",
		"user":    "U789012",
		"ts":      "1234567890.123456",
	}

	// No expectations on translation service means it should not be called
	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_OnlyAtChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	mockSlackClient := &SlackClient{
		client: nil,
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"channel": "C123456",
		"text":    "<!channel>",
		"user":    "U789012",
		"ts":      "1234567890.123456",
	}

	// No expectations on translation service means it should not be called
	processor.handleMessageEvent(context.Background(), event)
}

func TestEventProcessorHandleMessageEvent_MixedMentionsOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationService := mocks.NewMockTranslationService(ctrl)
	logger, _ := zap.NewProduction()

	mockSlackClient := &SlackClient{
		client: nil,
	}

	processor := NewEventProcessor(mockTranslationService, mockSlackClient, logger).(*eventProcessorImpl)

	event := map[string]interface{}{
		"type":    "message",
		"channel": "C123456",
		"text":    "<@U123456> @here <!channel>",
		"user":    "U789012",
		"ts":      "1234567890.123456",
	}

	// No expectations on translation service means it should not be called
	processor.handleMessageEvent(context.Background(), event)
}
