package domain

import "time"

type ChannelConfig struct {
	ID               string
	ChannelID        string
	AutoTranslate    bool
	SourceLanguages  []string
	TargetLanguage   string
	Enabled          bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ChannelRepository interface {
	Save(config *ChannelConfig) error
	GetByChannelID(channelID string) (*ChannelConfig, error)
	Update(config *ChannelConfig) error
	Delete(channelID string) error
	GetAll() ([]*ChannelConfig, error)
}
