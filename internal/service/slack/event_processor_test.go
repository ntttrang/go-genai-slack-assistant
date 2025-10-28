package slack

import (
	"context"
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

	// Assert that processor implements EventProcessor interface
	var _ EventProcessor = processor
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
