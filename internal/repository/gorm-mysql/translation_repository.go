package gormmysql

import (
	"fmt"

	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service"
	"gorm.io/gorm"
)

// TranslationRepositoryImpl implements service.TranslationRepository interface
type TranslationRepositoryImpl struct {
	db *gorm.DB
}

// NewTranslationRepository creates a new translation repository instance
func NewTranslationRepository(db *gorm.DB) service.TranslationRepository {
	return &TranslationRepositoryImpl{db: db}
}

func (tr *TranslationRepositoryImpl) Save(translation *model.Translation) error {
	if err := tr.db.Create(translation).Error; err != nil {
		return fmt.Errorf("failed to save translation: %w", err)
	}
	return nil
}

func (tr *TranslationRepositoryImpl) GetByHash(hash string) (*model.Translation, error) {
	translation := &model.Translation{}

	result := tr.db.Where("hash = ?", hash).First(translation)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get translation by hash: %w", result.Error)
	}

	return translation, nil
}

func (tr *TranslationRepositoryImpl) GetByID(id string) (*model.Translation, error) {
	translation := &model.Translation{}

	result := tr.db.Where("id = ?", id).First(translation)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get translation by id: %w", result.Error)
	}

	return translation, nil
}

func (tr *TranslationRepositoryImpl) GetByChannelID(channelID string, limit int) ([]*model.Translation, error) {
	var translations []*model.Translation

	result := tr.db.Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Find(&translations)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query translations: %w", result.Error)
	}

	return translations, nil
}
