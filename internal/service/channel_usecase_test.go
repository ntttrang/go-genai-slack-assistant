package service

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/testutils/mocks"
	"github.com/stretchr/testify/assert"
)

func TestChannelUseCaseCreateChannelConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockChannelRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	config := &model.ChannelConfig{
		ChannelID:       "C123",
		AutoTranslate:   true,
		SourceLanguages: []string{"en"},
		TargetLanguage:  "es",
		Enabled:         true,
		CreatedAt:       time.Now(),
	}

	mockRepo.EXPECT().Save(config).Return(nil)
	mockCache.EXPECT().Delete(gomock.Any()).Return(nil)

	useCase := NewChannelUseCase(mockRepo, mockCache)

	// Execute
	err := useCase.CreateChannelConfig(config)

	// Assert
	assert.NoError(t, err)
}

func TestChannelUseCaseGetChannelConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockChannelRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	expectedConfig := &model.ChannelConfig{
		ID:              "1",
		ChannelID:       "C123",
		AutoTranslate:   true,
		SourceLanguages: []string{"en"},
		TargetLanguage:  "es",
		Enabled:         true,
	}

	mockCache.EXPECT().Get("channel_config:C123").Return("", assert.AnError)
	mockRepo.EXPECT().GetByChannelID("C123").Return(expectedConfig, nil)
	mockCache.EXPECT().Set("channel_config:C123", "1", int64(3600)).Return(nil)

	useCase := NewChannelUseCase(mockRepo, mockCache)

	// Execute
	result, err := useCase.GetChannelConfig("C123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, result)
}

func TestChannelUseCaseDeleteChannelConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockChannelRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	mockRepo.EXPECT().Delete("C123").Return(nil)
	mockCache.EXPECT().Delete("channel_config:C123").Return(nil)

	useCase := NewChannelUseCase(mockRepo, mockCache)

	// Execute
	err := useCase.DeleteChannelConfig("C123")

	// Assert
	assert.NoError(t, err)
}

func TestChannelUseCaseIsChannelEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockChannelRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	enabledConfig := &model.ChannelConfig{
		ChannelID: "C123",
		Enabled:   true,
	}

	mockCache.EXPECT().Get("channel_config:C123").Return("", assert.AnError)
	mockRepo.EXPECT().GetByChannelID("C123").Return(enabledConfig, nil)
	mockCache.EXPECT().Set("channel_config:C123", "1", int64(3600)).Return(nil)

	useCase := NewChannelUseCase(mockRepo, mockCache)

	// Execute
	enabled, err := useCase.IsChannelEnabled("C123")

	// Assert
	assert.NoError(t, err)
	assert.True(t, enabled)
}

func TestChannelUseCaseImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockChannelRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	useCase := NewChannelUseCase(mockRepo, mockCache)

	// Assert that usecase implements ChannelService interface
	var _ ChannelService = useCase
	assert.NotNil(t, useCase)
}
