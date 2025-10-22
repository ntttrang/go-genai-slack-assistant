package usecase

import (
	"fmt"

	"github.com/ntttrang/python-genai-your-slack-assistant/domain"
)

type ChannelUseCase struct {
	repo domain.ChannelRepository
	cache domain.Cache
}

func NewChannelUseCase(repo domain.ChannelRepository, cache domain.Cache) *ChannelUseCase {
	return &ChannelUseCase{
		repo: repo,
		cache: cache,
	}
}

func (cu *ChannelUseCase) CreateChannelConfig(config *domain.ChannelConfig) error {
	if err := cu.repo.Save(config); err != nil {
		return fmt.Errorf("failed to create channel config: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("channel_config:%s", config.ChannelID)
	cu.cache.Delete(cacheKey)

	return nil
}

func (cu *ChannelUseCase) GetChannelConfig(channelID string) (*domain.ChannelConfig, error) {
	cacheKey := fmt.Sprintf("channel_config:%s", channelID)

	// Try cache first
	cachedJSON, err := cu.cache.Get(cacheKey)
	if err == nil && cachedJSON != "" {
		// For simplicity, skip cache deserialization
		// In production, would use proper JSON serialization
	}

	// Get from database
	config, err := cu.repo.GetByChannelID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel config: %w", err)
	}

	// Cache the result (1 hour TTL)
	cu.cache.Set(cacheKey, "1", 3600)

	return config, nil
}

func (cu *ChannelUseCase) UpdateChannelConfig(config *domain.ChannelConfig) error {
	if err := cu.repo.Update(config); err != nil {
		return fmt.Errorf("failed to update channel config: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("channel_config:%s", config.ChannelID)
	cu.cache.Delete(cacheKey)

	return nil
}

func (cu *ChannelUseCase) DeleteChannelConfig(channelID string) error {
	if err := cu.repo.Delete(channelID); err != nil {
		return fmt.Errorf("failed to delete channel config: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("channel_config:%s", channelID)
	cu.cache.Delete(cacheKey)

	return nil
}

func (cu *ChannelUseCase) ListAllChannelConfigs() ([]*domain.ChannelConfig, error) {
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
