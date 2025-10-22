package model

import "time"

type Translation struct {
	ID              string
	SourceMessageID string
	SourceText      string
	SourceLanguage  string
	TargetLanguage  string
	TranslatedText  string
	Hash            string
	UserID          string
	ChannelID       string
	CreatedAt       time.Time
	TTL             int64
}
