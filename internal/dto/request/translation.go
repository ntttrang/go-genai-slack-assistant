package request

import (
	"github.com/ntttrang/go-genai-slack-assistant/internal/dto"
)

type Translation struct {
	Text           string `json:"text" binding:"required"`
	SourceLanguage string `json:"source_language" binding:"required"`
	TargetLanguage string `json:"target_language" binding:"required"`
}

// Validate validates the translation request
func (t *Translation) Validate() *dto.Validator {
	v := dto.NewValidator()

	if t.Text == "" {
		v.Add("text", "text is required")
	} else if len(t.Text) > 5000 {
		v.Add("text", "text cannot exceed 5000 characters")
	}

	if t.SourceLanguage == "" {
		v.Add("source_language", "source_language is required")
	}

	if t.TargetLanguage == "" {
		v.Add("target_language", "target_language is required")
	}

	if t.SourceLanguage != "" && t.TargetLanguage != "" && t.SourceLanguage == t.TargetLanguage {
		v.Add("target_language", "source and target languages cannot be the same")
	}

	return v
}
