package service

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/middleware"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/testutils/mocks"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestTranslationUseCaseTranslate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTranslationRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockTranslator := mocks.NewMockTranslator(ctrl)

	// Setup expectations - cache miss
	mockCache.EXPECT().Get(gomock.Any()).Return("", assert.AnError)

	// Repo miss
	mockRepo.EXPECT().GetByHash(gomock.Any()).Return(nil, nil)

	// Translator succeeds
	mockTranslator.EXPECT().Translate("Hello", "en", "es").Return("Hola", nil)

	// Repo save succeeds
	mockRepo.EXPECT().Save(gomock.Any()).Return(nil)

	// Cache set succeeds
	mockCache.EXPECT().Set(gomock.Any(), "Hola", int64(3600)).Return(nil)

	securityMiddleware := createSecurityMiddleware()
	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600, securityMiddleware)
	assert.NotNil(t, useCase)
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

func createSecurityMiddleware() *middleware.SecurityMiddleware {
	inputValidator := security.NewInputValidator(5000)
	outputValidator := security.NewOutputValidator(10000)
	logger := zap.NewNop()
	return middleware.NewSecurityMiddleware(inputValidator, outputValidator, logger, true, true)
}

func TestTranslate_CacheHit(t *testing.T) {
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockCache)
	mockTranslator := new(MockTranslator)

	mockCache.On("Get", mock.Anything).Return("Hola", nil)

	securityMiddleware := createSecurityMiddleware()
	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	// Execute
	resp, err := useCase.Translate(request.Translation{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "es",
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Hola", resp.TranslatedText)
	assert.Equal(t, "Hello", resp.OriginalText)
	assert.Equal(t, "en", resp.SourceLanguage)
	assert.Equal(t, "es", resp.TargetLanguage)
}

func TestTranslationUseCaseTranslateFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTranslationRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockTranslator := mocks.NewMockTranslator(ctrl)

	// Setup expectations - cache hit
	mockCache.EXPECT().Get(gomock.Any()).Return("Cached Hola", nil)

	securityMiddleware := createSecurityMiddleware()
	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	// Execute
	resp, err := useCase.Translate(request.Translation{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "es",
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Cached Hola", resp.TranslatedText)
}

func TestTranslationUseCaseDetectLanguage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTranslationRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockTranslator := mocks.NewMockTranslator(ctrl)

	mockTranslator.EXPECT().DetectLanguage("Hello world").Return("en", nil)

	securityMiddleware := createSecurityMiddleware()
	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600, securityMiddleware)

	// Execute
	lang, err := useCase.DetectLanguage("Hello world")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "English", lang)
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

	securityMiddleware := createSecurityMiddleware()
	tu := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	req := request.Translation{
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

func TestTranslationUseCaseDetectLanguageVietnamese(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTranslationRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockTranslator := mocks.NewMockTranslator(ctrl)

	mockTranslator.EXPECT().DetectLanguage("Xin chào").Return("vi", nil)

	securityMiddleware := createSecurityMiddleware()
	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600, securityMiddleware)

	// Execute
	lang, err := useCase.DetectLanguage("Xin chào")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Vietnamese", lang)
}

func TestTranslationUseCaseImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTranslationRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockTranslator := mocks.NewMockTranslator(ctrl)

	securityMiddleware := createSecurityMiddleware()
	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600, securityMiddleware)

	// Assert that usecase implements TranslationService interface
	var _ TranslationService = useCase
	assert.NotNil(t, useCase)
}

func TestTranslate_AIError(t *testing.T) {
	mockRepo := new(MockTranslationRepository)
	mockCache := new(MockCache)
	mockTranslator := new(MockTranslator)

	mockCache.On("Get", mock.Anything).Return("", errors.New("not found"))
	mockRepo.On("GetByHash", mock.Anything).Return(nil, errors.New("not found"))
	mockTranslator.On("Translate", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("API error"))

	securityMiddleware := createSecurityMiddleware()
	tu := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 86400, securityMiddleware)

	req := request.Translation{
		Text:           "Hello",
		SourceLanguage: "en",
		TargetLanguage: "vi",
	}

	_, err := tu.Translate(req)

	assert.Error(t, err)
}
