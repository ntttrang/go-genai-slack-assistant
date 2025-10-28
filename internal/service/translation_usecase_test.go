package service

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/response"
	"github.com/ntttrang/go-genai-slack-assistant/internal/middleware"
	"github.com/ntttrang/go-genai-slack-assistant/internal/testutils/mocks"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/security"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupSecurityMiddleware() *middleware.SecurityMiddleware {
	inputValidator := security.NewInputValidator(5000)
	outputValidator := security.NewOutputValidator(10000)
	logger := zap.NewNop()
	return middleware.NewSecurityMiddleware(inputValidator, outputValidator, logger, true, true)
}

func TestTranslationUseCase_Translate(t *testing.T) {
	tests := []struct {
		name               string
		input              request.Translation
		cacheTTL           int64
		setupMocks         func(*mocks.MockCache, *mocks.MockTranslationRepository, *mocks.MockTranslator)
		expectedTranslated string
		expectedError      bool
		validateResponse   func(*testing.T, *response.Translation)
	}{
		{
			name: "cache hit",
			input: request.Translation{
				Text:           "Hello",
				SourceLanguage: "en",
				TargetLanguage: "es",
			},
			cacheTTL: 86400,
			setupMocks: func(cache *mocks.MockCache, repo *mocks.MockTranslationRepository, translator *mocks.MockTranslator) {
				cache.EXPECT().Get(gomock.Any()).Return("Hola", nil)
			},
			expectedTranslated: "Hola",
			expectedError:      false,
			validateResponse: func(t *testing.T, resp *response.Translation) {
				assert.Equal(t, "Hola", resp.TranslatedText)
				assert.Equal(t, "Hello", resp.OriginalText)
				assert.Equal(t, "en", resp.SourceLanguage)
				assert.Equal(t, "es", resp.TargetLanguage)
			},
		},
		{
			name: "cache miss - AI translation success",
			input: request.Translation{
				Text:           "Hello",
				SourceLanguage: "en",
				TargetLanguage: "es",
			},
			cacheTTL: 3600,
			setupMocks: func(cache *mocks.MockCache, repo *mocks.MockTranslationRepository, translator *mocks.MockTranslator) {
				cache.EXPECT().Get(gomock.Any()).Return("", errors.New("cache miss"))
				repo.EXPECT().GetByHash(gomock.Any()).Return(nil, nil)
				translator.EXPECT().Translate("Hello", "en", "es").Return("Hola", nil)
				repo.EXPECT().Save(gomock.Any()).Return(nil)
				cache.EXPECT().Set(gomock.Any(), "Hola", int64(3600)).Return(nil)
			},
			expectedTranslated: "Hola",
			expectedError:      false,
			validateResponse: func(t *testing.T, resp *response.Translation) {
				assert.Equal(t, "Hola", resp.TranslatedText)
				assert.Equal(t, "Hello", resp.OriginalText)
				assert.Equal(t, "en", resp.SourceLanguage)
				assert.Equal(t, "es", resp.TargetLanguage)
			},
		},
		{
			name: "Vietnamese translation",
			input: request.Translation{
				Text:           "Hello",
				SourceLanguage: "en",
				TargetLanguage: "vi",
			},
			cacheTTL: 86400,
			setupMocks: func(cache *mocks.MockCache, repo *mocks.MockTranslationRepository, translator *mocks.MockTranslator) {
				cache.EXPECT().Get(gomock.Any()).Return("", errors.New("not found"))
				repo.EXPECT().GetByHash(gomock.Any()).Return(nil, nil)
				translator.EXPECT().Translate("Hello", "en", "vi").Return("Xin chào", nil)
				repo.EXPECT().Save(gomock.Any()).Return(nil)
				cache.EXPECT().Set(gomock.Any(), "Xin chào", int64(86400)).Return(nil)
			},
			expectedTranslated: "Xin chào",
			expectedError:      false,
			validateResponse: func(t *testing.T, resp *response.Translation) {
				assert.Equal(t, "Xin chào", resp.TranslatedText)
			},
		},
		{
			name: "AI translation error",
			input: request.Translation{
				Text:           "Hello",
				SourceLanguage: "en",
				TargetLanguage: "vi",
			},
			cacheTTL: 86400,
			setupMocks: func(cache *mocks.MockCache, repo *mocks.MockTranslationRepository, translator *mocks.MockTranslator) {
				cache.EXPECT().Get(gomock.Any()).Return("", errors.New("not found"))
				repo.EXPECT().GetByHash(gomock.Any()).Return(nil, nil)
				translator.EXPECT().Translate(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("API error"))
			},
			expectedError: true,
		},
		{
			name: "cached translation with different text",
			input: request.Translation{
				Text:           "Goodbye",
				SourceLanguage: "en",
				TargetLanguage: "es",
			},
			cacheTTL: 86400,
			setupMocks: func(cache *mocks.MockCache, repo *mocks.MockTranslationRepository, translator *mocks.MockTranslator) {
				cache.EXPECT().Get(gomock.Any()).Return("Adiós", nil)
			},
			expectedTranslated: "Adiós",
			expectedError:      false,
			validateResponse: func(t *testing.T, resp *response.Translation) {
				assert.Equal(t, "Adiós", resp.TranslatedText)
				assert.Equal(t, "Goodbye", resp.OriginalText)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTranslationRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			mockTranslator := mocks.NewMockTranslator(ctrl)

			tt.setupMocks(mockCache, mockRepo, mockTranslator)

			securityMiddleware := setupSecurityMiddleware()
			useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, tt.cacheTTL, securityMiddleware)

			// Execute
			resp, err := useCase.Translate(tt.input)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTranslated, resp.TranslatedText)
				
				if tt.validateResponse != nil {
					tt.validateResponse(t, &resp)
				}
			}
		})
	}
}

func TestTranslationUseCase_DetectLanguage(t *testing.T) {
	tests := []struct {
		name             string
		inputText        string
		mockDetectedCode string
		mockError        error
		expectedLanguage string
		expectError      bool
	}{
		{
			name:             "detect English",
			inputText:        "Hello world",
			mockDetectedCode: "en",
			mockError:        nil,
			expectedLanguage: "English",
			expectError:      false,
		},
		{
			name:             "detect Vietnamese",
			inputText:        "Xin chào",
			mockDetectedCode: "vi",
			mockError:        nil,
			expectedLanguage: "Vietnamese",
			expectError:      false,
		},
		{
			name:             "detect Spanish",
			inputText:        "Hola mundo",
			mockDetectedCode: "es",
			mockError:        nil,
			expectedLanguage: "es",
			expectError:      false,
		},
		{
			name:             "detect French",
			inputText:        "Bonjour",
			mockDetectedCode: "fr",
			mockError:        nil,
			expectedLanguage: "fr",
			expectError:      false,
		},
		{
			name:             "detection error",
			inputText:        "???",
			mockDetectedCode: "",
			mockError:        errors.New("detection failed"),
			expectedLanguage: "",
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockTranslationRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			mockTranslator := mocks.NewMockTranslator(ctrl)

			mockTranslator.EXPECT().DetectLanguage(tt.inputText).Return(tt.mockDetectedCode, tt.mockError)

			securityMiddleware := setupSecurityMiddleware()
			useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600, securityMiddleware)

			// Execute
			lang, err := useCase.DetectLanguage(tt.inputText)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLanguage, lang)
			}
		})
	}
}

func TestTranslationUseCase_ImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTranslationRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockTranslator := mocks.NewMockTranslator(ctrl)

	securityMiddleware := setupSecurityMiddleware()
	useCase := NewTranslationUseCase(mockRepo, mockCache, mockTranslator, 3600, securityMiddleware)

	// Assert that usecase implements TranslationService interface
	var _ TranslationService = useCase
	assert.NotNil(t, useCase)
}
