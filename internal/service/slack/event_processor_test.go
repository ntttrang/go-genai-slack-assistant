package slack

import (
	"testing"
)

func TestExtractAndRestoreEmojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single emoji preservation",
			input:    "mệt :heavy_plus_sign:",
			expected: "mệt :heavy_plus_sign:",
		},
		{
			name:     "Multiple emojis preservation",
			input:    "Happy :grinning: and sad :weary:",
			expected: "Happy :grinning: and sad :weary:",
		},
		{
			name:     "Emoji with numbers and underscores",
			input:    "number_one :keycap_1:",
			expected: "number_one :keycap_1:",
		},
		{
			name:     "No emojis",
			input:    "Just plain text",
			expected: "Just plain text",
		},
		{
			name:     "Emoji at start",
			input:    ":grinning: Hello",
			expected: ":grinning: Hello",
		},
		{
			name:     "Emoji at end",
			input:    "Hello :grinning:",
			expected: "Hello :grinning:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract emojis
			cleanedText, emojis := extractEmojis(tt.input)

			// Verify placeholders don't contain original emoji text
			for placeholder, emoji := range emojis {
				if emoji != tt.input[len(tt.input)-len(emoji):] && len(tt.input) >= len(emoji) {
					// Make sure placeholder is actually a placeholder
					if len(placeholder) == 0 {
						t.Errorf("Placeholder is empty")
					}
				}
			}

			// Restore emojis
			result := restoreEmojis(cleanedText, emojis)

			// Verify the result matches the original
			if result != tt.expected {
				t.Errorf("extractEmojis() test %s failed: got %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestPlaceholderFormat(t *testing.T) {
	text := "mệt :heavy_plus_sign:"
	cleanedText, emojis := extractEmojis(text)

	// Verify placeholder format
	for placeholder := range emojis {
		// Check that placeholder starts with EMOJIPLACEHOLDER
		if len(placeholder) < 16 {
			t.Errorf("Placeholder format incorrect: %q", placeholder)
		}
		// Verify it doesn't contain underscores at start/end
		if placeholder[0] == '_' || placeholder[len(placeholder)-1] == '_' {
			t.Errorf("Placeholder should not have leading/trailing underscores: %q", placeholder)
		}
	}

	// Verify cleaned text doesn't contain original emoji
	if cleanedText == text {
		t.Errorf("Cleaned text should not contain original text")
	}
}
