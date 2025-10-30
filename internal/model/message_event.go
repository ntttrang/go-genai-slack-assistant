package model

import "time"

// MessageEvent represents a Slack event to be processed
type MessageEvent struct {
	EventID    string
	ChannelID  string
	UserID     string
	MessageTS  string
	Payload    map[string]interface{}
	ReceivedAt time.Time
	Sequence   uint64
}

// GetQueueKey returns the key for queue management
// Using channel_id ensures ordering at channel level for all messages in the channel
func (e *MessageEvent) GetQueueKey() string {
	return e.ChannelID
}
