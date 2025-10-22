package repository

import (
	"github.com/ntttrang/python-genai-your-slack-assistant/internal/model"
)

type TranslationRepository interface {
	Save(translation *model.Translation) error
	GetByHash(hash string) (*model.Translation, error)
	GetByID(id string) (*model.Translation, error)
	GetByChannelID(channelID string, limit int) ([]*model.Translation, error)
}
