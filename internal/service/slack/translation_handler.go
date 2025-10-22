package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/ntttrang/python-genai-your-slack-assistant/internal/dto/request"
	svc "github.com/ntttrang/python-genai-your-slack-assistant/internal/service"
	"github.com/ntttrang/python-genai-your-slack-assistant/pkg/language"
	"go.uber.org/zap"
)

type Translator interface {
	Translate(text, sourceLanguage, targetLanguage string) (string, error)
	DetectLanguage(text string) (string, error)
}

type TranslationHandler struct {
	translationUseCase *svc.TranslationUseCase
	slackClient        *SlackClient
	languageDetector   *language.LanguageDetector
	logger             *zap.Logger
}

func NewTranslationHandler(
	translationUseCase *svc.TranslationUseCase,
	slackClient *SlackClient,
	languageDetector *language.LanguageDetector,
	logger *zap.Logger,
) *TranslationHandler {
	return &TranslationHandler{
		translationUseCase: translationUseCase,
		slackClient:        slackClient,
		languageDetector:   languageDetector,
		logger:             logger,
	}
}

func (th *TranslationHandler) TranslateAndPostReply(
	ctx context.Context,
	channelID string,
	messageText string,
	threadTS string,
	targetLanguage string,
) error {
	if messageText == "" {
		return fmt.Errorf("empty message text")
	}

	// Detect source language
	detectedLang, err := th.languageDetector.DetectLanguage(messageText)
	if err != nil {
		th.logger.Error("Failed to detect language", zap.Error(err))
		return err
	}

	sourceLangCode, err := th.languageDetector.GetLanguageCode(detectedLang)
	if err != nil {
		th.logger.Error("Failed to get language code", zap.Error(err))
		return err
	}

	// Skip if source language is same as target
	if sourceLangCode == targetLanguage {
		th.logger.Info("Source and target languages are same, skipping translation",
			zap.String("language", sourceLangCode))
		return nil
	}

	// Translate
	req := request.Translation{
		Text:           messageText,
		SourceLanguage: detectedLang,
		TargetLanguage: th.getLanguageName(targetLanguage),
	}

	resp, err := th.translationUseCase.Translate(req)
	if err != nil {
		th.logger.Error("Translation failed", zap.Error(err))
		return err
	}

	// Post to thread
	responseText := fmt.Sprintf("🇻🇳 *Vietnamese Translation:*\n%s", resp.TranslatedText)
	_, _, err = th.slackClient.PostMessage(channelID, responseText, threadTS)
	if err != nil {
		th.logger.Error("Failed to post message", zap.Error(err))
		return err
	}

	th.logger.Info("Translation posted successfully",
		zap.String("channel_id", channelID),
		zap.String("thread_ts", threadTS))

	return nil
}

func (th *TranslationHandler) getLanguageName(code string) string {
	langMap := map[string]string{
		"en": "English",
		"vi": "Vietnamese",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"zh": "Chinese",
		"ja": "Japanese",
		"ko": "Korean",
	}

	if name, ok := langMap[code]; ok {
		return name
	}
	return "Unknown"
}

func (th *TranslationHandler) SanitizeText(text string) string {
	text = strings.TrimSpace(text)
	if len(text) > 10240 {
		text = text[:10240]
	}
	return text
}
