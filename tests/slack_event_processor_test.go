package tests

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/middleware"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/security"
)

type MockTranslator struct {
	mock.Mock
}

func (m *MockTranslator) Translate(text, sourceLanguage, targetLanguage string) (string, error) {
	args := m.Called(text, sourceLanguage, targetLanguage)
	return args.String(0), args.Error(1)
}

func (m *MockTranslator) DetectLanguage(text string) (string, error) {
	args := m.Called(text)
	return args.String(0), args.Error(1)
}

type MockTranslationRepository struct {
	mock.Mock
}

func (m *MockTranslationRepository) Save(translation *model.Translation) error {
	args := m.Called(translation)
	return args.Error(0)
}

func (m *MockTranslationRepository) GetByHash(hash string) (*model.Translation, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Translation), args.Error(1)
}

func (m *MockTranslationRepository) GetByID(id string) (*model.Translation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Translation), args.Error(1)
}

func (m *MockTranslationRepository) GetByChannelID(channelID string, limit int) ([]*model.Translation, error) {
	args := m.Called(channelID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Translation), args.Error(1)
}

type MockRedisCache struct {
	mock.Mock
}

func (m *MockRedisCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisCache) Set(key string, value string, ttl int64) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockRedisCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockRedisCache) Exists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

// TestVietnameseMessageToEnglishTranslation tests the use case for Vietnamese message translation
func TestVietnameseMessageToEnglishTranslation(t *testing.T) {
	mockTranslator := new(MockTranslator)
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockRedisCache)

	vietnameseMessage := "Xin chào, bạn khỏe không?"
	englishTranslation := "Hello, how are you?"

	// Mock cache miss
	mockCache.On("Get", mock.Anything).Return("", errors.New("cache miss"))

	// Mock database miss
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("not found"))

	// Mock translation call - Vietnamese to English
	mockTranslator.On("Translate", vietnameseMessage, "Vietnamese", "English").
		Return(englishTranslation, nil)

	// Mock save to database
	mockRepo.On("Save", mock.Anything).Return(nil)

	// Mock cache set
	mockCache.On("Set", mock.Anything, englishTranslation, int64(86400)).Return(nil)

	// Create translation use case with security middleware
	inputValidator := security.NewInputValidator(5000)
	outputValidator := security.NewOutputValidator(10000)
	logger := zap.NewNop()
	securityMiddleware := middleware.NewSecurityMiddleware(inputValidator, outputValidator, logger, true, true)
	tu := service.NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	// Test translation Vietnamese to English
	req := request.Translation{
		Text:           vietnameseMessage,
		SourceLanguage: "Vietnamese",
		TargetLanguage: "English",
	}

	result, err := tu.Translate(req)

	assert.NoError(t, err)
	assert.Equal(t, englishTranslation, result.TranslatedText)
	assert.Equal(t, "Vietnamese", result.SourceLanguage)
	assert.Equal(t, "English", result.TargetLanguage)

	// Verify translator was called
	mockTranslator.AssertCalled(t, "Translate", vietnameseMessage, "Vietnamese", "English")
}

// TestEnglishMessageToVietnameseTranslation tests the use case for English message translation
func TestEnglishMessageToVietnameseTranslation(t *testing.T) {
	mockTranslator := new(MockTranslator)
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockRedisCache)

	englishMessage := "Hello, how are you?"
	vietnameseTranslation := "Xin chào, bạn khỏe không?"

	// Mock cache miss
	mockCache.On("Get", mock.Anything).Return("", errors.New("cache miss"))

	// Mock database miss
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("not found"))

	// Mock translation call - English to Vietnamese
	mockTranslator.On("Translate", englishMessage, "English", "Vietnamese").
		Return(vietnameseTranslation, nil)

	// Mock save to database
	mockRepo.On("Save", mock.Anything).Return(nil)

	// Mock cache set
	mockCache.On("Set", mock.Anything, vietnameseTranslation, int64(86400)).Return(nil)

	// Create translation use case with security middleware
	inputValidator := security.NewInputValidator(5000)
	outputValidator := security.NewOutputValidator(10000)
	logger := zap.NewNop()
	securityMiddleware := middleware.NewSecurityMiddleware(inputValidator, outputValidator, logger, true, true)
	tu := service.NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	// Test translation English to Vietnamese
	req := request.Translation{
		Text:           englishMessage,
		SourceLanguage: "English",
		TargetLanguage: "Vietnamese",
	}

	result, err := tu.Translate(req)

	assert.NoError(t, err)
	assert.Equal(t, vietnameseTranslation, result.TranslatedText)
	assert.Equal(t, "English", result.SourceLanguage)
	assert.Equal(t, "Vietnamese", result.TargetLanguage)

	// Verify translator was called
	mockTranslator.AssertCalled(t, "Translate", englishMessage, "English", "Vietnamese")
}

// TestTranslationUseCaseIntegration tests the full translation flow
func TestTranslationUseCaseIntegration(t *testing.T) {
	mockTranslator := new(MockTranslator)
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockRedisCache)

	// Mock all the calls
	mockCache.On("Get", mock.Anything).Return("", errors.New("cache miss"))
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("not found"))
	mockTranslator.On("Translate", "Hello", "English", "Vietnamese").Return("Xin chào", nil)
	mockRepo.On("Save", mock.Anything).Return(nil)
	mockCache.On("Set", mock.Anything, "Xin chào", int64(86400)).Return(nil)

	// Create use case with security middleware
	inputValidator := security.NewInputValidator(5000)
	outputValidator := security.NewOutputValidator(10000)
	logger := zap.NewNop()
	securityMiddleware := middleware.NewSecurityMiddleware(inputValidator, outputValidator, logger, true, true)
	tu := service.NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	// Test translation
	req := request.Translation{
		Text:           "Hello",
		SourceLanguage: "English",
		TargetLanguage: "Vietnamese",
	}

	result, err := tu.Translate(req)

	assert.NoError(t, err)
	assert.Equal(t, "Xin chào", result.TranslatedText)
	assert.Equal(t, "English", result.SourceLanguage)
	assert.Equal(t, "Vietnamese", result.TargetLanguage)

	// Verify all mocks were called
	mockTranslator.AssertCalled(t, "Translate", "Hello", "English", "Vietnamese")
	mockRepo.AssertCalled(t, "Save", mock.Anything)
}

// Test translation flow with payload marshaling
func TestEventPayloadParsing(t *testing.T) {
	payload := map[string]interface{}{
		"type": "event_callback",
		"event": map[string]interface{}{
			"type":    "message",
			"channel": "C123456",
			"user":    "U123456",
			"text":    "Xin chào",
			"ts":      "1234567890.123456",
		},
	}

	// Marshal and unmarshal to simulate Slack JSON payload
	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var parsedPayload map[string]interface{}
	err = json.Unmarshal(data, &parsedPayload)
	assert.NoError(t, err)

	assert.Equal(t, "event_callback", parsedPayload["type"])
	event := parsedPayload["event"].(map[string]interface{})
	assert.Equal(t, "message", event["type"])
	assert.Equal(t, "Xin chào", event["text"])
}
