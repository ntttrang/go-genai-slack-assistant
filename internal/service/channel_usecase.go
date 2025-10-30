package service

import (
	"fmt"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
)

// ChannelRepository defines the interface for channel configuration persistence.
// This interface is owned by the ChannelUseCase and defined where it's consumed.
type ChannelRepository interface {
	Save(config *model.ChannelConfig) error
	GetByChannelID(channelID string) (*model.ChannelConfig, error)
	Update(config *model.ChannelConfig) error
	Delete(channelID string) error
	GetAll() ([]*model.ChannelConfig, error)
}

var _ ChannelService = (*ChannelUseCase)(nil)

type ChannelUseCase struct {
	repo  ChannelRepository
	cache Cache
}

func NewChannelUseCase(repo ChannelRepository, cache Cache) *ChannelUseCase {
	return &ChannelUseCase{
		repo:  repo,
		cache: cache,
	}
}

func (cu *ChannelUseCase) CreateChannelConfig(config *model.ChannelConfig) error {
	if err := cu.repo.Save(config); err != nil {
		return fmt.Errorf("failed to create channel config: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("channel_config:%s", config.ChannelID)
	_ = cu.cache.Delete(cacheKey)

	return nil
}

func (cu *ChannelUseCase) GetChannelConfig(channelID string) (*model.ChannelConfig, error) {
	cacheKey := fmt.Sprintf("channel_config:%s", channelID)

	// Try cache first
	_, _ = cu.cache.Get(cacheKey)

	// Get from database
	config, err := cu.repo.GetByChannelID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel config: %w", err)
	}

	// Cache the result (1 hour TTL)
	cacheValue := "0"
	if config.Enabled {
		cacheValue = "1"
	}
	_ = cu.cache.Set(cacheKey, cacheValue, 3600)

	return config, nil
}

func (cu *ChannelUseCase) UpdateChannelConfig(config *model.ChannelConfig) error {
	if err := cu.repo.Update(config); err != nil {
		return fmt.Errorf("failed to update channel config: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("channel_config:%s", config.ChannelID)
	_ = cu.cache.Delete(cacheKey)

	return nil
}

func (cu *ChannelUseCase) DeleteChannelConfig(channelID string) error {
	if err := cu.repo.Delete(channelID); err != nil {
		return fmt.Errorf("failed to delete channel config: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("channel_config:%s", channelID)
	_ = cu.cache.Delete(cacheKey)

	return nil
}

func (cu *ChannelUseCase) ListAllChannelConfigs() ([]*model.ChannelConfig, error) {
	configs, err := cu.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list channel configs: %w", err)
	}

	return configs, nil
}

func (cu *ChannelUseCase) IsChannelEnabled(channelID string) (bool, error) {
	config, err := cu.GetChannelConfig(channelID)
	if err != nil {
		// If config doesn't exist, default to enabled
		return true, nil
	}

	return config.Enabled, nil
}
