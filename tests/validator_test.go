package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ntttrang/python-genai-your-slack-assistant/internal/model"
)

func TestValidateMessage_Empty(t *testing.T) {
	v := model.NewMessageValidator()
	err := v.ValidateMessage("")

	assert.Error(t, err)
	assert.IsType(t, &model.DomainError{}, err)
	assert.Equal(t, model.ErrorTypeValidation, err.(*model.DomainError).Type)
}

func TestValidateMessage_TooLong(t *testing.T) {
	v := model.NewMessageValidator()
	longMessage := strings.Repeat("a", model.MaxMessageLength+1)
	err := v.ValidateMessage(longMessage)

	assert.Error(t, err)
	assert.IsType(t, &model.DomainError{}, err)
	assert.Equal(t, model.ErrorTypeValidation, err.(*model.DomainError).Type)
}

func TestValidateMessage_Valid(t *testing.T) {
	v := model.NewMessageValidator()
	err := v.ValidateMessage("Hello World")

	assert.NoError(t, err)
}

func TestValidateChannelID_Empty(t *testing.T) {
	v := model.NewMessageValidator()
	err := v.ValidateChannelID("")

	assert.Error(t, err)
}

func TestValidateChannelID_Invalid(t *testing.T) {
	v := model.NewMessageValidator()
	err := v.ValidateChannelID("invalid-channel")

	assert.Error(t, err)
}

func TestValidateChannelID_Valid(t *testing.T) {
	v := model.NewMessageValidator()
	err := v.ValidateChannelID("C123456789")

	assert.NoError(t, err)
}

func TestValidateUserID_Empty(t *testing.T) {
	v := model.NewMessageValidator()
	err := v.ValidateUserID("")

	assert.Error(t, err)
}

func TestValidateUserID_Valid(t *testing.T) {
	v := model.NewMessageValidator()
	err := v.ValidateUserID("U123456789")

	assert.NoError(t, err)
}

func TestSanitizeInput_RemovesNullBytes(t *testing.T) {
	v := model.NewMessageValidator()
	input := "Hello\x00World"
	sanitized := v.SanitizeInput(input)

	assert.Equal(t, "HelloWorld", sanitized)
}

func TestSanitizeInput_TrimsWhitespace(t *testing.T) {
	v := model.NewMessageValidator()
	input := "  Hello World  "
	sanitized := v.SanitizeInput(input)

	assert.Equal(t, "Hello World", sanitized)
}

func TestSanitizeInput_RemovesControlCharacters(t *testing.T) {
	v := model.NewMessageValidator()
	input := "Hello\x01World"
	sanitized := v.SanitizeInput(input)

	assert.Equal(t, "HelloWorld", sanitized)
}

func TestSanitizeInput_PreservesNewlines(t *testing.T) {
	v := model.NewMessageValidator()
	input := "Hello\nWorld"
	sanitized := v.SanitizeInput(input)

	assert.Equal(t, "Hello\nWorld", sanitized)
}
