package gormmysql

import (
	"fmt"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
	"gorm.io/gorm"
)

// ChannelRepositoryImpl implements service.ChannelRepository interface
type ChannelRepositoryImpl struct {
	db *gorm.DB
}

// NewChannelRepository creates a new channel repository instance
func NewChannelRepository(db *gorm.DB) service.ChannelRepository {
	return &ChannelRepositoryImpl{db: db}
}

func (cr *ChannelRepositoryImpl) Save(config *model.ChannelConfig) error {
	if err := cr.db.Create(config).Error; err != nil {
		return fmt.Errorf("failed to save channel config: %w", err)
	}
	return nil
}

func (cr *ChannelRepositoryImpl) GetByChannelID(channelID string) (*model.ChannelConfig, error) {
	config := &model.ChannelConfig{}

	result := cr.db.Where("channel_id = ?", channelID).First(config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel config not found")
		}
		return nil, fmt.Errorf("failed to get channel config: %w", result.Error)
	}

	return config, nil
}

func (cr *ChannelRepositoryImpl) Update(config *model.ChannelConfig) error {
	result := cr.db.Model(&model.ChannelConfig{}).Where("channel_id = ?", config.ChannelID).Updates(map[string]interface{}{
		"auto_translate":   config.AutoTranslate,
		"source_languages": config.SourceLanguages,
		"target_language":  config.TargetLanguage,
		"enabled":          config.Enabled,
		"updated_at":       config.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update channel config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("channel config not found")
	}

	return nil
}

func (cr *ChannelRepositoryImpl) Delete(channelID string) error {
	result := cr.db.Where("channel_id = ?", channelID).Delete(&model.ChannelConfig{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete channel config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("channel config not found")
	}

	return nil
}

func (cr *ChannelRepositoryImpl) GetAll() ([]*model.ChannelConfig, error) {
	var configs []*model.ChannelConfig

	result := cr.db.Order("created_at DESC").Find(&configs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query channel configs: %w", result.Error)
	}

	return configs, nil
}
