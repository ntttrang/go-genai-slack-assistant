package service

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/testutils/mocks"
	"github.com/stretchr/testify/assert"
)

func TestChannelUseCase(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(*testing.T, *mocks.MockChannelRepository, *mocks.MockCache, ChannelService)
	}{
		{
			name: "create channel config",
			testFunc: func(t *testing.T, mockRepo *mocks.MockChannelRepository, mockCache *mocks.MockCache, useCase ChannelService) {
				config := &model.ChannelConfig{
					ChannelID:       "C123",
					AutoTranslate:   true,
					SourceLanguages: `["en"]`,
					TargetLanguage:  "es",
					Enabled:         true,
					CreatedAt:       time.Now(),
				}

				mockRepo.EXPECT().Save(config).Return(nil)
				mockCache.EXPECT().Delete(gomock.Any()).Return(nil)

				err := useCase.CreateChannelConfig(config)

				assert.NoError(t, err)
			},
		},
		{
			name: "get channel config",
			testFunc: func(t *testing.T, mockRepo *mocks.MockChannelRepository, mockCache *mocks.MockCache, useCase ChannelService) {
				expectedConfig := &model.ChannelConfig{
					ID:              "1",
					ChannelID:       "C123",
					AutoTranslate:   true,
					SourceLanguages: `["en"]`,
					TargetLanguage:  "es",
					Enabled:         true,
				}

				mockCache.EXPECT().Get("channel_config:C123").Return("", assert.AnError)
				mockRepo.EXPECT().GetByChannelID("C123").Return(expectedConfig, nil)
				mockCache.EXPECT().Set("channel_config:C123", gomock.Any(), int64(3600)).Return(nil)

				result, err := useCase.GetChannelConfig("C123")

				assert.NoError(t, err)
				assert.Equal(t, expectedConfig, result)
			},
		},
		{
			name: "delete channel config",
			testFunc: func(t *testing.T, mockRepo *mocks.MockChannelRepository, mockCache *mocks.MockCache, useCase ChannelService) {
				mockRepo.EXPECT().Delete("C123").Return(nil)
				mockCache.EXPECT().Delete("channel_config:C123").Return(nil)

				err := useCase.DeleteChannelConfig("C123")

				assert.NoError(t, err)
			},
		},
		{
			name: "is channel enabled",
			testFunc: func(t *testing.T, mockRepo *mocks.MockChannelRepository, mockCache *mocks.MockCache, useCase ChannelService) {
				enabledConfig := &model.ChannelConfig{
					ChannelID: "C123",
					Enabled:   true,
				}

				mockCache.EXPECT().Get("channel_config:C123").Return("", assert.AnError)
				mockRepo.EXPECT().GetByChannelID("C123").Return(enabledConfig, nil)
				mockCache.EXPECT().Set("channel_config:C123", "1", int64(3600)).Return(nil)

				enabled, err := useCase.IsChannelEnabled("C123")

				assert.NoError(t, err)
				assert.True(t, enabled)
			},
		},
		{
			name: "is channel disabled",
			testFunc: func(t *testing.T, mockRepo *mocks.MockChannelRepository, mockCache *mocks.MockCache, useCase ChannelService) {
				disabledConfig := &model.ChannelConfig{
					ChannelID: "C456",
					Enabled:   false,
				}

				mockCache.EXPECT().Get("channel_config:C456").Return("", assert.AnError)
				mockRepo.EXPECT().GetByChannelID("C456").Return(disabledConfig, nil)
				mockCache.EXPECT().Set("channel_config:C456", "0", int64(3600)).Return(nil)

				enabled, err := useCase.IsChannelEnabled("C456")

				assert.NoError(t, err)
				assert.False(t, enabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockChannelRepository(ctrl)
			mockCache := mocks.NewMockCache(ctrl)
			useCase := NewChannelUseCase(mockRepo, mockCache)

			tt.testFunc(t, mockRepo, mockCache, useCase)
		})
	}
}

func TestChannelUseCaseImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockChannelRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	useCase := NewChannelUseCase(mockRepo, mockCache)

	var _ ChannelService = useCase
	assert.NotNil(t, useCase)
}
