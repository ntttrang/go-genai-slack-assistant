package service

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/testutils/mocks"
	"github.com/stretchr/testify/assert"
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

	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600)

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

	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600)

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

	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600)

	// Execute
	lang, err := useCase.DetectLanguage("Hello world")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "English", lang)
}

func TestTranslationUseCaseDetectLanguageVietnamese(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTranslationRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockTranslator := mocks.NewMockTranslator(ctrl)

	mockTranslator.EXPECT().DetectLanguage("Xin chào").Return("vi", nil)

	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600)

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

	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600)

	// Assert that usecase implements TranslationService interface
	var _ TranslationService = useCase
	assert.NotNil(t, useCase)
}
