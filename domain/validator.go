package domain

import (
	"fmt"
	"regexp"
	"strings"
)

const MaxMessageLength = 10 * 1024 // 10KB

type Validator interface {
	ValidateMessage(msg string) error
	ValidateChannelID(channelID string) error
	ValidateUserID(userID string) error
	SanitizeInput(input string) string
}

type MessageValidator struct{}

func NewMessageValidator() *MessageValidator {
	return &MessageValidator{}
}

func (v *MessageValidator) ValidateMessage(msg string) error {
	msg = strings.TrimSpace(msg)

	if msg == "" {
		return NewValidationError("message cannot be empty")
	}

	if len(msg) > MaxMessageLength {
		return NewValidationError(fmt.Sprintf("message exceeds maximum length of %d bytes", MaxMessageLength))
	}

	return nil
}

func (v *MessageValidator) ValidateChannelID(channelID string) error {
	if channelID == "" {
		return NewValidationError("channel ID cannot be empty")
	}

	// Slack channel IDs start with C or are direct message IDs
	if !isValidSlackID(channelID) {
		return NewValidationError("invalid channel ID format")
	}

	return nil
}

func (v *MessageValidator) ValidateUserID(userID string) error {
	if userID == "" {
		return NewValidationError("user ID cannot be empty")
	}

	if !isValidSlackID(userID) {
		return NewValidationError("invalid user ID format")
	}

	return nil
}

func (v *MessageValidator) SanitizeInput(input string) string {
	// Remove null bytes to prevent prompt injection
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove control characters except newline and tab
	input = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(input, "")

	return input
}

func isValidSlackID(id string) bool {
	// Slack IDs are alphanumeric strings, typically 9-11 characters
	return len(id) > 0 && regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(id)
}
