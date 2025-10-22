package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ntttrang/python-genai-your-slack-assistant/internal/model"
)

type TranslationUseCase struct {
	repo       model.TranslationRepository
	cache      model.Cache
	translator model.Translator
	cacheTTL   int64
}

func NewTranslationUseCase(
	repo model.TranslationRepository,
	cache model.Cache,
	translator model.Translator,
	cacheTTL int64,
) *TranslationUseCase {
	return &TranslationUseCase{
		repo:       repo,
		cache:      cache,
		translator: translator,
		cacheTTL:   cacheTTL,
	}
}

func (tu *TranslationUseCase) Translate(req model.TranslationRequest) (model.TranslationResponse, error) {
	hash := tu.generateHash(req.Text, req.SourceLanguage, req.TargetLanguage)
	cacheKey := fmt.Sprintf("translation:%s", hash)

	// Try to get from cache
	cachedResult, err := tu.cache.Get(cacheKey)
	if err == nil && cachedResult != "" {
		return model.TranslationResponse{
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
		return model.TranslationResponse{
			OriginalText:   req.Text,
			TranslatedText: existingTranslation.TranslatedText,
			SourceLanguage: req.SourceLanguage,
			TargetLanguage: req.TargetLanguage,
		}, nil
	}

	// Call AI to translate
	translatedText, err := tu.translator.Translate(req.Text, req.SourceLanguage, req.TargetLanguage)
	if err != nil {
		return model.TranslationResponse{}, fmt.Errorf("translation failed: %w", err)
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
		return model.TranslationResponse{}, fmt.Errorf("failed to save translation: %w", err)
	}

	// Store in cache
	tu.cache.Set(cacheKey, translatedText, tu.cacheTTL)

	return model.TranslationResponse{
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
