package repository

import (
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
)

type MessageRepository interface {
	Save(msg *model.Message) error
	GetByID(id string) (*model.Message, error)
}
