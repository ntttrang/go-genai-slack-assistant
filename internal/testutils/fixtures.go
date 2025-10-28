package testutils

import (
	"time"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
)

// Test Fixtures Factory
// Provides reusable test data creation functions to reduce boilerplate in tests

// NewTestTranslation creates a test Translation entity with default values
func NewTestTranslation() *model.Translation {
	return &model.Translation{
		ID:              "test-translation-1",
		SourceMessageID: "test-msg-123",
		SourceText:      "Hello world",
		SourceLanguage:  "English",
		TargetLanguage:  "Vietnamese",
		TranslatedText:  "Xin chào thế giới",
		Hash:            "test-hash-abc123",
		UserID:          "U12345",
		ChannelID:       "C12345",
		CreatedAt:       time.Now(),
		TTL:             3600,
	}
}

// NewTestTranslationWithOptions creates a Translation with custom fields
func NewTestTranslationWithOptions(opts TranslationOptions) *model.Translation {
	t := NewTestTranslation()
	
	if opts.ID != "" {
		t.ID = opts.ID
	}
	if opts.SourceText != "" {
		t.SourceText = opts.SourceText
	}
	if opts.TranslatedText != "" {
		t.TranslatedText = opts.TranslatedText
	}
	if opts.SourceLanguage != "" {
		t.SourceLanguage = opts.SourceLanguage
	}
	if opts.TargetLanguage != "" {
		t.TargetLanguage = opts.TargetLanguage
	}
	if opts.Hash != "" {
		t.Hash = opts.Hash
	}
	if opts.ChannelID != "" {
		t.ChannelID = opts.ChannelID
	}
	if opts.UserID != "" {
		t.UserID = opts.UserID
	}
	
	return t
}

type TranslationOptions struct {
	ID              string
	SourceText      string
	TranslatedText  string
	SourceLanguage  string
	TargetLanguage  string
	Hash            string
	ChannelID       string
	UserID          string
}

// NewTestChannelConfig creates a test ChannelConfig entity with default values
func NewTestChannelConfig() *model.ChannelConfig {
	return &model.ChannelConfig{
		ID:              "test-channel-config-1",
		ChannelID:       "C12345",
		Enabled:         true,
		SourceLanguages: []string{"English", "Japanese"},
		TargetLanguage:  "Vietnamese",
		AutoTranslate:   true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// NewTestChannelConfigWithOptions creates a ChannelConfig with custom fields
func NewTestChannelConfigWithOptions(opts ChannelConfigOptions) *model.ChannelConfig {
	c := NewTestChannelConfig()
	
	if opts.ID != "" {
		c.ID = opts.ID
	}
	if opts.ChannelID != "" {
		c.ChannelID = opts.ChannelID
	}
	if opts.Enabled != nil {
		c.Enabled = *opts.Enabled
	}
	if opts.SourceLanguages != nil {
		c.SourceLanguages = opts.SourceLanguages
	}
	if opts.TargetLanguage != "" {
		c.TargetLanguage = opts.TargetLanguage
	}
	if opts.AutoTranslate != nil {
		c.AutoTranslate = *opts.AutoTranslate
	}
	
	return c
}

type ChannelConfigOptions struct {
	ID              string
	ChannelID       string
	Enabled         *bool
	SourceLanguages []string
	TargetLanguage  string
	AutoTranslate   *bool
}

// NewTestMessage creates a test Message entity with default values
func NewTestMessage() *model.Message {
	return &model.Message{
		ID:        "test-message-1",
		ChannelID: "C12345",
		UserID:    "U12345",
		Text:      "Test message",
		Timestamp: "1234567890.123456",
		ThreadTs:  "",
	}
}

// NewTestMessageWithOptions creates a Message with custom fields
func NewTestMessageWithOptions(opts MessageOptions) *model.Message {
	m := NewTestMessage()
	
	if opts.ID != "" {
		m.ID = opts.ID
	}
	if opts.ChannelID != "" {
		m.ChannelID = opts.ChannelID
	}
	if opts.UserID != "" {
		m.UserID = opts.UserID
	}
	if opts.Text != "" {
		m.Text = opts.Text
	}
	if opts.Timestamp != "" {
		m.Timestamp = opts.Timestamp
	}
	
	return m
}

type MessageOptions struct {
	ID        string
	ChannelID string
	UserID    string
	Text      string
	Timestamp string
}

// Helper functions for common test scenarios

// NewTranslationsForChannel creates multiple translations for the same channel
func NewTranslationsForChannel(channelID string, count int) []*model.Translation {
	translations := make([]*model.Translation, count)
	for i := 0; i < count; i++ {
		translations[i] = NewTestTranslationWithOptions(TranslationOptions{
			ID:        generateTestID(i),
			ChannelID: channelID,
		})
	}
	return translations
}

// NewEnabledChannelConfig creates an enabled channel config
func NewEnabledChannelConfig(channelID string) *model.ChannelConfig {
	enabled := true
	return NewTestChannelConfigWithOptions(ChannelConfigOptions{
		ChannelID: channelID,
		Enabled:   &enabled,
	})
}

// NewDisabledChannelConfig creates a disabled channel config
func NewDisabledChannelConfig(channelID string) *model.ChannelConfig {
	disabled := false
	return NewTestChannelConfigWithOptions(ChannelConfigOptions{
		ChannelID: channelID,
		Enabled:   &disabled,
	})
}

func generateTestID(index int) string {
	return time.Now().Format("20060102150405") + "-" + string(rune('0'+index))
}
