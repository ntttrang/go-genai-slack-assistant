package model

import (
	"testing"
	"time"
)

func TestTranslationCreation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		translation *Translation
		validate    func(*testing.T, *Translation)
	}{
		{
			name: "all fields set correctly",
			translation: &Translation{
				ID:              "test-1",
				SourceText:      "Hello",
				SourceLanguage:  "en",
				TargetLanguage:  "es",
				TranslatedText:  "Hola",
				Hash:            "test-hash",
				ChannelID:       "C123",
				UserID:          "U456",
				SourceMessageID: "M789",
				CreatedAt:       now,
				TTL:             3600,
			},
			validate: func(t *testing.T, trans *Translation) {
				if trans.ID != "test-1" {
					t.Errorf("expected ID test-1, got %s", trans.ID)
				}
				if trans.SourceText != "Hello" {
					t.Errorf("expected source text Hello, got %s", trans.SourceText)
				}
				if trans.TranslatedText != "Hola" {
					t.Errorf("expected translated text Hola, got %s", trans.TranslatedText)
				}
				if trans.SourceLanguage != "en" {
					t.Errorf("expected source language en, got %s", trans.SourceLanguage)
				}
				if trans.TargetLanguage != "es" {
					t.Errorf("expected target language es, got %s", trans.TargetLanguage)
				}
			},
		},
		{
			name: "Vietnamese translation",
			translation: &Translation{
				ID:              "test-2",
				SourceText:      "Good morning",
				SourceLanguage:  "en",
				TargetLanguage:  "vi",
				TranslatedText:  "Chào buổi sáng",
				Hash:            "test-hash-2",
				ChannelID:       "C456",
				UserID:          "U789",
				SourceMessageID: "M012",
				CreatedAt:       now,
				TTL:             7200,
			},
			validate: func(t *testing.T, trans *Translation) {
				if trans.ID != "test-2" {
					t.Errorf("expected ID test-2, got %s", trans.ID)
				}
				if trans.TargetLanguage != "vi" {
					t.Errorf("expected target language vi, got %s", trans.TargetLanguage)
				}
				if trans.TTL != 7200 {
					t.Errorf("expected TTL 7200, got %d", trans.TTL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.translation)
		})
	}
}

func TestChannelConfigCreation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		config   *ChannelConfig
		validate func(*testing.T, *ChannelConfig)
	}{
		{
			name: "enabled config with auto translate",
			config: &ChannelConfig{
				ID:              "1",
				ChannelID:       "C123",
				AutoTranslate:   true,
				SourceLanguages: `["en", "vi"]`,
				TargetLanguage:  "es",
				Enabled:         true,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			validate: func(t *testing.T, config *ChannelConfig) {
				if config.ChannelID != "C123" {
					t.Errorf("expected channel ID C123, got %s", config.ChannelID)
				}
				if !config.AutoTranslate {
					t.Error("expected AutoTranslate to be true")
				}
				if !config.Enabled {
					t.Error("expected Enabled to be true")
				}
				if config.SourceLanguages != `["en", "vi"]` {
					t.Errorf("expected source languages [\"en\", \"vi\"], got %s", config.SourceLanguages)
				}
			},
		},
		{
			name: "disabled config",
			config: &ChannelConfig{
				ID:              "2",
				ChannelID:       "C456",
				AutoTranslate:   false,
				SourceLanguages: `["fr"]`,
				TargetLanguage:  "en",
				Enabled:         false,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			validate: func(t *testing.T, config *ChannelConfig) {
				if config.Enabled {
					t.Error("expected Enabled to be false")
				}
				if config.AutoTranslate {
					t.Error("expected AutoTranslate to be false")
				}
				if config.TargetLanguage != "en" {
					t.Errorf("expected target language en, got %s", config.TargetLanguage)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.config)
		})
	}
}
