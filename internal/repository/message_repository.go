package repository

import (
	"github.com/ntttrang/python-genai-your-slack-assistant/internal/model"
)

type MessageRepository interface {
	Save(msg *model.Message) error
	GetByID(id string) (*model.Message, error)
}
