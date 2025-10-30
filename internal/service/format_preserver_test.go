package service

import (
	"testing"
)

func TestFormatPreserver_Emojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single emoji",
			input:    "Hello :smile: world",
			expected: "Hello :smile: world",
		},
		{
			name:     "Multiple emojis",
			input:    ":wave: Hello :smile: :tada:",
			expected: ":wave: Hello :smile: :tada:",
		},
		{
			name:     "Emoji at start",
			input:    ":fire: This is hot",
			expected: ":fire: This is hot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preserver := NewFormatPreserver()
			cleaned := preserver.Extract(tt.input)
			restored := preserver.Restore(cleaned)

			if restored != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, restored)
			}
		})
	}
}

func TestFormatPreserver_CodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Inline code",
			input:    "Use `make` command",
			expected: "Use `make` command",
		},
		{
			name:     "Triple backtick code block",
			input:    "```\nfunc Hello() {\n}\n```",
			expected: "```\nfunc Hello() {\n}\n```",
		},
		{
			name:     "Mixed code",
			input:    "Run `npm start` or see ```config.js```",
			expected: "Run `npm start` or see ```config.js```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preserver := NewFormatPreserver()
			cleaned := preserver.Extract(tt.input)
			restored := preserver.Restore(cleaned)

			if restored != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, restored)
			}
		})
	}
}

func TestFormatPreserver_Links(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL link",
			input:    "Check https://github.com for more",
			expected: "Check https://github.com for more",
		},
		{
			name:     "Slack link",
			input:    "Message <@U12345678>",
			expected: "Message <@U12345678>",
		},
		{
			name:     "Channel link",
			input:    "Join <#C12345678|general>",
			expected: "Join <#C12345678|general>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preserver := NewFormatPreserver()
			cleaned := preserver.Extract(tt.input)
			restored := preserver.Restore(cleaned)

			if restored != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, restored)
			}
		})
	}
}

func TestFormatPreserver_LineBreaks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single line",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Multiple lines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Line with emoji and break",
			input:    ":smile: Hello\nWorld :wave:",
			expected: ":smile: Hello\nWorld :wave:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preserver := NewFormatPreserver()
			cleaned := preserver.Extract(tt.input)
			restored := preserver.Restore(cleaned)

			if restored != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, restored)
			}
		})
	}
}

func TestFormatPreserver_BulletPoints(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple bullet list",
			input:    "* Item 1\n* Item 2",
			expected: "* Item 1\n* Item 2",
		},
		{
			name:     "Bullet list with indentation",
			input:    " * Test 1\n  * Test 2",
			expected: " * Test 1\n  * Test 2",
		},
		{
			name:     "Dash bullet list",
			input:    "- Item A\n- Item B",
			expected: "- Item A\n- Item B",
		},
		{
			name:     "Numbered list",
			input:    "1. First\n2. Second\n3. Third",
			expected: "1. First\n2. Second\n3. Third",
		},
		{
			name:     "Numbered list with indentation",
			input:    "  1. Nested one\n  2. Nested two",
			expected: "  1. Nested one\n  2. Nested two",
		},
		{
			name:     "Mixed bullet and numbered",
			input:    "* Main\n  1. Sub one\n  2. Sub two",
			expected: "* Main\n  1. Sub one\n  2. Sub two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preserver := NewFormatPreserver()
			cleaned := preserver.Extract(tt.input)
			restored := preserver.Restore(cleaned)

			if restored != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, restored)
			}
		})
	}
}

func TestFormatPreserver_Combined(t *testing.T) {
	input := ":wave: Hello world\nCheck `npm start` at https://example.com\n:smile: Done <#C12345>"
	expected := input

	preserver := NewFormatPreserver()
	cleaned := preserver.Extract(input)
	restored := preserver.Restore(cleaned)

	if restored != expected {
		t.Errorf("expected %q, got %q", expected, restored)
	}
}

func TestFormatPreserver_BulletPointsWithOtherFormats(t *testing.T) {
	input := " * Test 1\n  * Test 2 with :smile:\n  * Test 3 with `code`"
	expected := input

	preserver := NewFormatPreserver()
	cleaned := preserver.Extract(input)
	restored := preserver.Restore(cleaned)

	if restored != expected {
		t.Errorf("expected %q, got %q", expected, restored)
	}
}

func TestFormatPreserver_ExtractOnly(t *testing.T) {
	input := "Hello :smile: with `code` and https://link.com"
	preserver := NewFormatPreserver()
	cleaned := preserver.Extract(input)

	// Verify that emojis, code, and links are replaced with placeholders
	if cleaned == input {
		t.Errorf("Extract should have replaced patterns but didn't")
	}

	// Verify placeholders exist
	if len(preserver.emojis) == 0 || len(preserver.codeBlocks) == 0 || len(preserver.links) == 0 {
		t.Errorf("expected patterns to be extracted, got emojis=%d, codes=%d, links=%d",
			len(preserver.emojis), len(preserver.codeBlocks), len(preserver.links))
	}
}

func TestFormatPreserver_Reset(t *testing.T) {
	preserver := NewFormatPreserver()
	preserver.Extract("Hello :smile: and `code`")

	if len(preserver.emojis) == 0 || len(preserver.codeBlocks) == 0 {
		t.Fatal("Extract should have stored patterns")
	}

	preserver.Reset()

	if len(preserver.emojis) != 0 || len(preserver.codeBlocks) != 0 {
		t.Errorf("Reset should clear patterns but got emojis=%d, codes=%d",
			len(preserver.emojis), len(preserver.codeBlocks))
	}
}

func TestFormatPreserver_ConvertUserMentions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mappings map[string]string
		expected string
	}{
		{
			name:     "Message with single user mention",
			input:    "Hello <@U12345678> how are you?",
			mappings: map[string]string{"U12345678": "john"},
			expected: "Hello john how are you?",
		},
		{
			name:     "Message with multiple user mentions",
			input:    "Hi <@U12345678> and <@U87654321>, let's talk",
			mappings: map[string]string{"U12345678": "alice", "U87654321": "bob"},
			expected: "Hi alice and bob, let's talk",
		},
		{
			name:     "Message with user mention and other content",
			input:    "<@U12345678> check this :smile: https://example.com",
			mappings: map[string]string{"U12345678": "carol"},
			expected: "carol check this :smile: https://example.com",
		},
		{
			name:     "Message without user mentions",
			input:    "Hello world, this is a message",
			mappings: map[string]string{},
			expected: "Hello world, this is a message",
		},
		{
			name:     "User mention with no mapping (fallback to ID)",
			input:    "<@U12345678> check this",
			mappings: map[string]string{},
			expected: "U12345678 check this",
		},
		{
			name:     "User mention with channel reference",
			input:    "<@U12345678> <#C87654321|general>",
			mappings: map[string]string{"U12345678": "dave"},
			expected: "dave <#C87654321|general>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preserver := NewFormatPreserver()
			preserver.SetUsernameMappings(tt.mappings)
			cleaned := preserver.Extract(tt.input)
			restored := preserver.RestoreWithOptions(cleaned, true)

			if restored != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, restored)
			}
		})
	}
}

func TestFormatPreserver_RestoreWithOptions_KeepMentions(t *testing.T) {
	input := "Hey <@U12345678>, check this"
	expected := input

	preserver := NewFormatPreserver()
	cleaned := preserver.Extract(input)
	restored := preserver.RestoreWithOptions(cleaned, false)

	if restored != expected {
		t.Errorf("expected %q, got %q", expected, restored)
	}
}
