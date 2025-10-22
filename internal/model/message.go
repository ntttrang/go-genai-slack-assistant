package model

type Message struct {
	ID        string
	UserID    string
	ChannelID string
	Text      string
	Timestamp string
	ThreadTs  string
}

type MessageRepository interface {
	Save(msg *Message) error
	GetByID(id string) (*Message, error)
}
