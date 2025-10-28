package service

import (
	"context"

	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/request"
	"github.com/ntttrang/go-genai-slack-assistant/internal/dto/response"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
)

// TranslationService defines the interface for translation use cases
type TranslationService interface {
	Translate(req request.Translation) (response.Translation, error)
	DetectLanguage(text string) (string, error)
}

// ChannelService defines the interface for channel configuration use cases
type ChannelService interface {
	CreateChannelConfig(config *model.ChannelConfig) error
	GetChannelConfig(channelID string) (*model.ChannelConfig, error)
	UpdateChannelConfig(config *model.ChannelConfig) error
	DeleteChannelConfig(channelID string) error
	ListAllChannelConfigs() ([]*model.ChannelConfig, error)
	IsChannelEnabled(channelID string) (bool, error)
}

// EventProcessorService defines the interface for event processing
type EventProcessorService interface {
	ProcessEvent(ctx context.Context, payload map[string]interface{})
}
