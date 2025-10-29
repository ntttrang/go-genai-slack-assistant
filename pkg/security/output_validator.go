package security

import (
	"regexp"
	"strings"
)

type OutputValidator struct {
	maxOutputLength int
}

type OutputValidationResult struct {
	IsValid     bool
	CleanedText string
	Issues      []string
}

func NewOutputValidator(maxLength int) *OutputValidator {
	return &OutputValidator{
		maxOutputLength: maxLength,
	}
}

func (v *OutputValidator) ValidateTranslation(output, originalInput string) OutputValidationResult {
	result := OutputValidationResult{
		IsValid:     true,
		CleanedText: output,
		Issues:      []string{},
	}

	if v.containsSystemPromptLeakage(output) {
		result.Issues = append(result.Issues, "Output contains system prompt leakage")
		result.IsValid = false
	}

	// if v.containsSuspiciousAcknowledgment(output) {
	// 	result.Issues = append(result.Issues, "Output contains acknowledgment of injected instructions")
	// 	result.IsValid = false
	// }

	if len(output) > v.maxOutputLength {
		result.Issues = append(result.Issues, "Output exceeds maximum length")
		result.IsValid = false
	}

	if len(strings.TrimSpace(output)) < len(originalInput)/10 {
		result.Issues = append(result.Issues, "Output suspiciously short")
		result.IsValid = false
	}

	result.CleanedText = v.cleanOutput(output)

	return result
}

func (v *OutputValidator) containsSystemPromptLeakage(output string) bool {
	lowerOutput := strings.ToLower(output)
	leakagePatterns := []string{
		"you are a",
		"i am a translation",
		"as an ai",
		"i cannot",
		"i must not",
		"my instructions",
		"system prompt",
		"<userinput>",
		"</userinput>",
	}

	for _, pattern := range leakagePatterns {
		if strings.Contains(lowerOutput, pattern) {
			return true
		}
	}

	return false
}

// func (v *OutputValidator) containsSuspiciousAcknowledgment(output string) bool {
// 	lowerOutput := strings.ToLower(output)
// 	acknowledgments := []string{
// 		"i understand",
// 		"i will now",
// 		"okay, i'll",
// 		"sure, i can",
// 		"acknowledged",
// 	}

// 	for _, ack := range acknowledgments {
// 		if strings.Contains(lowerOutput, ack) {
// 			return true
// 		}
// 	}

// 	return false
// }

func (v *OutputValidator) cleanOutput(output string) string {
	cleaned := strings.TrimSpace(output)
	cleaned = regexp.MustCompile(`(?i)(translation|translated text):\s*`).ReplaceAllString(cleaned, "")
	cleaned = regexp.MustCompile(`^["']|["']$`).ReplaceAllString(cleaned, "")
	return cleaned
}
