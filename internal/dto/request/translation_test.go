package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslationValidate_ValidRequest(t *testing.T) {
	req := &Translation{
		Text:           "Hello world",
		SourceLanguage: "English",
		TargetLanguage: "Vietnamese",
	}

	v := req.Validate()

	assert.True(t, v.Valid())
	assert.Empty(t, v.Errors())
}

func TestTranslationValidate_EmptyText(t *testing.T) {
	req := &Translation{
		Text:           "",
		SourceLanguage: "English",
		TargetLanguage: "Vietnamese",
	}

	v := req.Validate()

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 1)
	assert.Equal(t, "text", v.Errors()[0].Field)
	assert.Equal(t, "text is required", v.Errors()[0].Message)
}

func TestTranslationValidate_TextTooLong(t *testing.T) {
	req := &Translation{
		Text:           strings.Repeat("a", 5001),
		SourceLanguage: "English",
		TargetLanguage: "Vietnamese",
	}

	v := req.Validate()

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 1)
	assert.Equal(t, "text", v.Errors()[0].Field)
	assert.Equal(t, "text cannot exceed 5000 characters", v.Errors()[0].Message)
}

func TestTranslationValidate_EmptySourceLanguage(t *testing.T) {
	req := &Translation{
		Text:           "Hello",
		SourceLanguage: "",
		TargetLanguage: "Vietnamese",
	}

	v := req.Validate()

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 1)
	assert.Equal(t, "source_language", v.Errors()[0].Field)
}

func TestTranslationValidate_EmptyTargetLanguage(t *testing.T) {
	req := &Translation{
		Text:           "Hello",
		SourceLanguage: "English",
		TargetLanguage: "",
	}

	v := req.Validate()

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 1)
	assert.Equal(t, "target_language", v.Errors()[0].Field)
}

func TestTranslationValidate_SameSourceAndTarget(t *testing.T) {
	req := &Translation{
		Text:           "Hello",
		SourceLanguage: "English",
		TargetLanguage: "English",
	}

	v := req.Validate()

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 1)
	assert.Equal(t, "target_language", v.Errors()[0].Field)
	assert.Equal(t, "source and target languages cannot be the same", v.Errors()[0].Message)
}

func TestTranslationValidate_MaxTextLength(t *testing.T) {
	req := &Translation{
		Text:           strings.Repeat("a", 5000),
		SourceLanguage: "English",
		TargetLanguage: "Vietnamese",
	}

	v := req.Validate()

	assert.True(t, v.Valid())
	assert.Empty(t, v.Errors())
}

func TestTranslationValidate_MultipleErrors(t *testing.T) {
	req := &Translation{
		Text:           "",
		SourceLanguage: "",
		TargetLanguage: "",
	}

	v := req.Validate()

	assert.False(t, v.Valid())
	assert.Len(t, v.Errors(), 3)
}
