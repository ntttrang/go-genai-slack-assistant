package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/response"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/translator"
)

// TranslationRepository defines the interface for translation persistence.
// This interface is owned by the TranslationUseCase and defined where it's consumed.
type TranslationRepository interface {
	Save(translation *model.Translation) error
	GetByHash(hash string) (*model.Translation, error)
	GetByID(id string) (*model.Translation, error)
	GetByChannelID(channelID string, limit int) ([]*model.Translation, error)
}

var _ TranslationService = (*TranslationUseCase)(nil)

type TranslationUseCase struct {
	repo       TranslationRepository
	cache      Cache
	translator translator.Translator
	cacheTTL   int64
}

func NewTranslationUseCase(
	repo TranslationRepository,
	cache Cache,
	translator translator.Translator,
	cacheTTL int64,
) *TranslationUseCase {
	return &TranslationUseCase{
		repo:       repo,
		cache:      cache,
		translator: translator,
		cacheTTL:   cacheTTL,
	}
}

func (tu *TranslationUseCase) Translate(req request.Translation) (response.Translation, error) {
	hash := tu.generateHash(req.Text, req.SourceLanguage, req.TargetLanguage)
	cacheKey := fmt.Sprintf("translation:%s", hash)

	// Try to get from cache
	cachedResult, err := tu.cache.Get(cacheKey)
	if err == nil && cachedResult != "" {
		return response.Translation{
			OriginalText:   req.Text,
			TranslatedText: cachedResult,
			SourceLanguage: req.SourceLanguage,
			TargetLanguage: req.TargetLanguage,
		}, nil
	}

	// Try to get from database
	existingTranslation, err := tu.repo.GetByHash(hash)
	if err == nil && existingTranslation != nil {
		tu.cache.Set(cacheKey, existingTranslation.TranslatedText, tu.cacheTTL)
		return response.Translation{
			OriginalText:   req.Text,
			TranslatedText: existingTranslation.TranslatedText,
			SourceLanguage: req.SourceLanguage,
			TargetLanguage: req.TargetLanguage,
		}, nil
	}

	// Call AI to translate
	translatedText, err := tu.translator.Translate(req.Text, req.SourceLanguage, req.TargetLanguage)
	if err != nil {
		return response.Translation{}, fmt.Errorf("translation failed: %w", err)
	}

	// Store in database
	translation := &model.Translation{
		ID:             generateID(),
		SourceText:     req.Text,
		SourceLanguage: req.SourceLanguage,
		TargetLanguage: req.TargetLanguage,
		TranslatedText: translatedText,
		Hash:           hash,
		CreatedAt:      time.Now(),
		TTL:            tu.cacheTTL,
	}

	if err := tu.repo.Save(translation); err != nil {
		return response.Translation{}, fmt.Errorf("failed to save translation: %w", err)
	}

	// Store in cache
	tu.cache.Set(cacheKey, translatedText, tu.cacheTTL)

	return response.Translation{
		OriginalText:   req.Text,
		TranslatedText: translatedText,
		SourceLanguage: req.SourceLanguage,
		TargetLanguage: req.TargetLanguage,
	}, nil
}

func (tu *TranslationUseCase) generateHash(text, sourceLang, targetLang string) string {
	h := sha256.New()
	h.Write([]byte(text + sourceLang + targetLang))
	return hex.EncodeToString(h.Sum(nil))
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (tu *TranslationUseCase) DetectLanguage(text string) (string, error) {
	langCode, err := tu.translator.DetectLanguage(text)
	if err != nil {
		return "", fmt.Errorf("language detection failed: %w", err)
	}
	return normalizeLanguageCode(langCode), nil
}

func normalizeLanguageCode(code string) string {
	code = strings.TrimSpace(code)
	switch code {
	case "en", "EN", "english", "eng":
		return "English"
	case "vi", "VI", "vietnamese", "vie":
		return "Vietnamese"
	default:
		if code == "English" || code == "Vietnamese" {
			return code
		}
		return code
	}
}
