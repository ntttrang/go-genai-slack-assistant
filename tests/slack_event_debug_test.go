package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/middleware"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/security"
)

// TestTranslationFlowEnglishToVietnamese tests the full translation scenario
func TestTranslationFlowEnglishToVietnamese(t *testing.T) {
	// Setup mocks
	mockTranslator := new(MockTranslator)
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockRedisCache)

	englishMessage := "Hello, How are you?"
	vietnameseTranslation := "Xin chào bạn khỏe không?"

	// Mock translation
	mockTranslator.On("Translate", mock.Anything, "English", "Vietnamese").
		Return(vietnameseTranslation, nil)

	// Mock cache miss
	mockCache.On("Get", mock.Anything).Return("", errors.New("cache miss"))

	// Mock database miss
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("record not found"))

	// Mock save
	mockRepo.On("Save", mock.Anything).Return(nil)

	// Mock cache set
	mockCache.On("Set", mock.Anything, vietnameseTranslation, int64(86400)).Return(nil)

	// Create translation use case with security middleware
	inputValidator := security.NewInputValidator(5000)
	outputValidator := security.NewOutputValidator(10000)
	logger := zap.NewNop()
	securityMiddleware := middleware.NewSecurityMiddleware(inputValidator, outputValidator, logger, true, true)
	tu := service.NewTranslationUseCase(logger, mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	// Test: Translate English message
	result, err := tu.Translate(request.Translation{
		Text:           englishMessage,
		SourceLanguage: "English",
		TargetLanguage: "Vietnamese",
	})

	assert.NoError(t, err)
	assert.Equal(t, vietnameseTranslation, result.TranslatedText)

	t.Logf("✓ Full English→Vietnamese flow works!")
	t.Logf("  Input: %s", englishMessage)
	t.Logf("  Output: %s", vietnameseTranslation)
}

// TestTranslationFlowVietnameseToEnglish tests Vietnamese to English scenario
func TestTranslationFlowVietnameseToEnglish(t *testing.T) {
	// Setup mocks
	mockTranslator := new(MockTranslator)
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockRedisCache)

	vietnameseMessage := "Xin chào bạn khỏe không?"
	englishTranslation := "Hello, how are you?"

	// Mock translation
	mockTranslator.On("Translate", mock.Anything, "Vietnamese", "English").
		Return(englishTranslation, nil)

	// Mock cache miss
	mockCache.On("Get", mock.Anything).Return("", errors.New("cache miss"))

	// Mock database miss
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("not found"))

	// Mock save
	mockRepo.On("Save", mock.Anything).Return(nil)

	// Mock cache set
	mockCache.On("Set", mock.Anything, englishTranslation, int64(86400)).Return(nil)

	// Create translation use case with security middleware
	inputValidator := security.NewInputValidator(5000)
	outputValidator := security.NewOutputValidator(10000)
	logger := zap.NewNop()
	securityMiddleware := middleware.NewSecurityMiddleware(inputValidator, outputValidator, logger, true, true)
	tu := service.NewTranslationUseCase(logger, mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	// Test: Translate Vietnamese message
	result, err := tu.Translate(request.Translation{
		Text:           vietnameseMessage,
		SourceLanguage: "Vietnamese",
		TargetLanguage: "English",
	})

	assert.NoError(t, err)
	assert.Equal(t, englishTranslation, result.TranslatedText)

	t.Logf("✓ Full Vietnamese→English flow works!")
	t.Logf("  Input: %s", vietnameseMessage)
	t.Logf("  Output: %s", englishTranslation)
}
