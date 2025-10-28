package slack

import "context"

// EventProcessor defines the interface for Slack event processing
type EventProcessor interface {
	ProcessEvent(ctx context.Context, payload map[string]interface{})
}
