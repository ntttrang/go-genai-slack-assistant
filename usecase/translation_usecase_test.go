package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ntttrang/python-genai-your-slack-assistant/domain"
)

type MockTranslationRepository struct {
	mock.Mock
}

func (m *MockTranslationRepository) Save(translation *domain.Translation) error {
	args := m.Called(translation)
	return args.Error(0)
}

func (m *MockTranslationRepository) GetByHash(hash string) (*domain.Translation, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Translation), args.Error(1)
}

func (m *MockTranslationRepository) GetByID(id string) (*domain.Translation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Translation), args.Error(1)
}

func (m *MockTranslationRepository) GetByChannelID(channelID string, limit int) ([]*domain.Translation, error) {
	args := m.Called(channelID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Translation), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(key string, value string, ttl int64) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) Exists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

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

func TestTranslate_CacheHit(t *testing.T) {
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockCache)
	mockTranslator := new(MockTranslator)

	mockCache.On("Get", mock.Anything).Return("translated text", nil)

	tu := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400)

	req := domain.TranslationRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "vi",
	}

	result, err := tu.Translate(req)

	assert.NoError(t, err)
	assert.Equal(t, "translated text", result.TranslatedText)
	mockCache.AssertCalled(t, "Get", mock.Anything)
	mockTranslator.AssertNotCalled(t, "Translate", mock.Anything, mock.Anything, mock.Anything)
}

func TestTranslate_DatabaseHit(t *testing.T) {
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockCache)
	mockTranslator := new(MockTranslator)

	mockCache.On("Get", mock.Anything).Return("", errors.New("not found"))
	mockRepo.On("GetByHash", mock.Anything).Return(&domain.Translation{
		TranslatedText: "translated from db",
	}, nil)
	mockCache.On("Set", mock.Anything, "translated from db", int64(86400)).Return(nil)

	tu := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400)

	req := domain.TranslationRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "vi",
	}

	result, err := tu.Translate(req)

	assert.NoError(t, err)
	assert.Equal(t, "translated from db", result.TranslatedText)
	mockTranslator.AssertNotCalled(t, "Translate", mock.Anything, mock.Anything, mock.Anything)
}

func TestTranslate_AITranslation(t *testing.T) {
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockCache)
	mockTranslator := new(MockTranslator)

	mockCache.On("Get", mock.Anything).Return("", errors.New("not found"))
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("not found"))
	mockTranslator.On("Translate", "Hello", "en", "vi").Return("Xin chào", nil)
	mockRepo.On("Save", mock.Anything).Return(nil)
	mockCache.On("Set", mock.Anything, "Xin chào", int64(86400)).Return(nil)

	tu := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400)

	req := domain.TranslationRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "vi",
	}

	result, err := tu.Translate(req)

	assert.NoError(t, err)
	assert.Equal(t, "Xin chào", result.TranslatedText)
	mockTranslator.AssertCalled(t, "Translate", "Hello", "en", "vi")
	mockRepo.AssertCalled(t, "Save", mock.Anything)
	mockCache.AssertCalled(t, "Set", mock.Anything, "Xin chào", int64(86400))
}

func TestTranslate_AIError(t *testing.T) {
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockCache)
	mockTranslator := new(MockTranslator)

	mockCache.On("Get", mock.Anything).Return("", errors.New("not found"))
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("not found"))
	mockTranslator.On("Translate", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("API error"))

	tu := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400)

	req := domain.TranslationRequest{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "vi",
	}

	_, err := tu.Translate(req)

	assert.Error(t, err)
	mockTranslator.AssertCalled(t, "Translate", mock.Anything, mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "Save", mock.Anything)
}
