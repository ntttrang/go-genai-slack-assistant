package model

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
