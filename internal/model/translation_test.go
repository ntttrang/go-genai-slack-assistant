package model

import (
	"testing"
	"time"
)

func TestTranslationCreation(t *testing.T) {
	trans := &Translation{
		ID:               "test-1",
		SourceText:       "Hello",
		SourceLanguage:   "en",
		TargetLanguage:   "es",
		TranslatedText:   "Hola",
		Hash:             "test-hash",
		ChannelID:        "C123",
		UserID:           "U456",
		SourceMessageID:  "M789",
		CreatedAt:        time.Now(),
		TTL:              3600,
	}

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
}

func TestChannelConfigCreation(t *testing.T) {
	config := &ChannelConfig{
		ID:              "1",
		ChannelID:       "C123",
		AutoTranslate:   true,
		SourceLanguages: []string{"en", "vi"},
		TargetLanguage:  "es",
		Enabled:         true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if config.ChannelID != "C123" {
		t.Errorf("expected channel ID C123, got %s", config.ChannelID)
	}

	if !config.AutoTranslate {
		t.Error("expected AutoTranslate to be true")
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}

	if len(config.SourceLanguages) != 2 {
		t.Errorf("expected 2 source languages, got %d", len(config.SourceLanguages))
	}
}
