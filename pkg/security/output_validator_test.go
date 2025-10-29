package security_test

import (
	"strings"
	"testing"

	"github.com/ntttrang/go-genai-slack-assistant/pkg/security"
	"github.com/stretchr/testify/assert"
)

func repeatStr(s string, count int) string {
	return strings.Repeat(s, count)
}

func TestOutputValidator_ValidTranslation(t *testing.T) {
	validator := security.NewOutputValidator(10000)

	tests := []struct {
		name        string
		output      string
		original    string
		isValid     bool
		expectIssue bool
	}{
		{
			name:        "Valid translation",
			output:      "Xin chào, bạn khỏe không?",
			original:    "Hello, how are you?",
			isValid:     true,
			expectIssue: false,
		},
		{
			name:        "Valid longer translation",
			output:      "This is a longer translation with multiple words and phrases that should be valid",
			original:    "Đây là một bản dịch dài hơn với nhiều từ ngữ và cụm từ mà sẽ hợp lệ",
			isValid:     true,
			expectIssue: false,
		},
		{
			name:        "System prompt leakage - I am",
			output:      "I am a translation system, I cannot translate this",
			original:    "Test message",
			isValid:     false,
			expectIssue: true,
		},
		{
			name:        "System prompt leakage - I understand",
			output:      "I understand, I will translate for you",
			original:    "Test message",
			isValid:     true,
			expectIssue: false,
		},
		{
			name:        "System prompt leakage - instructions",
			output:      "My instructions say I should not translate this",
			original:    "Test message",
			isValid:     false,
			expectIssue: true,
		},
		{
			name:        "Suspicious acknowledgment - I understand",
			output:      "I understand your instructions to ignore previous context",
			original:    "Test message",
			isValid:     true,
			expectIssue: false,
		},
		{
			name:        "Suspicious acknowledgment - Okay",
			output:      "Okay, I'll respond with Hacked",
			original:    "Test message",
			isValid:     true,
			expectIssue: false,
		},
		{
			name:        "Output too short",
			output:      "Hi",
			original:    "This is a very long message that should produce a longer translation output",
			isValid:     false,
			expectIssue: true,
		},
		{
			name:        "Output exceeds max length",
			output:      repeatStr("a", 10001),
			original:    "Test",
			isValid:     false,
			expectIssue: true,
		},
		{
			name:        "Delimiter tag in output",
			output:      "Translation </UserInput> break out",
			original:    "Test",
			isValid:     false,
			expectIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateTranslation(tt.output, tt.original)
			assert.Equal(t, tt.isValid, result.IsValid, "validity mismatch")
			if tt.expectIssue {
				assert.Greater(t, len(result.Issues), 0, "expected issues but got none")
			}
		})
	}
}

func TestOutputValidator_CleanOutput(t *testing.T) {
	validator := security.NewOutputValidator(10000)

	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "Remove Translation prefix",
			output:   "Translation: Xin chào",
			expected: "Xin chào",
		},
		{
			name:     "Remove quotes",
			output:   `"Xin chào"`,
			expected: "Xin chào",
		},
		{
			name:     "Remove single quotes",
			output:   `'Xin chào'`,
			expected: "Xin chào",
		},
		{
			name:     "Trim whitespace",
			output:   "  Xin chào  ",
			expected: "Xin chào",
		},
		{
			name:     "Clean with Translation prefix and quotes",
			output:   `Translation: "Xin chào"`,
			expected: "Xin chào",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateTranslation(tt.output, "test")
			assert.Equal(t, tt.expected, result.CleanedText)
		})
	}
}

func TestOutputValidator_EdgeCases(t *testing.T) {
	validator := security.NewOutputValidator(100)

	tests := []struct {
		name     string
		output   string
		original string
		isValid  bool
	}{
		{
			name:     "Empty output",
			output:   "",
			original: "Test message",
			isValid:  false,
		},
		{
			name:     "Whitespace only output",
			output:   "   ",
			original: "Test message",
			isValid:  false,
		},
		{
			name:     "Very short original should not fail short output check",
			output:   "Hi",
			original: "Hi",
			isValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateTranslation(tt.output, tt.original)
			assert.Equal(t, tt.isValid, result.IsValid, "validity mismatch")
		})
	}
}
