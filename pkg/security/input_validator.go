package security

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type ThreatLevel int

const (
	ThreatLevelNone ThreatLevel = iota
	ThreatLevelLow
	ThreatLevelMedium
	ThreatLevelHigh
	ThreatLevelCritical
)

type InputValidator struct {
	maxLength          int
	suspiciousPatterns []*regexp.Regexp
	blockList          []string
}

type ValidationResult struct {
	IsValid          bool
	SanitizedText    string
	Warnings         []string
	ThreatLevel      ThreatLevel
	DetectedPatterns []string
}

func NewInputValidator(maxLength int) *InputValidator {
	return &InputValidator{
		maxLength:          maxLength,
		suspiciousPatterns: compileSuspiciousPatterns(),
		blockList:          loadBlockList(),
	}
}

func (v *InputValidator) Validate(text string) ValidationResult {
	result := ValidationResult{
		IsValid:          true,
		SanitizedText:    text,
		Warnings:         []string{},
		ThreatLevel:      ThreatLevelNone,
		DetectedPatterns: []string{},
	}

	// 0. Check for delimiter tag injection (CRITICAL)
	if v.containsDelimiterTags(text) {
		result.IsValid = false
		result.ThreatLevel = ThreatLevelCritical
		result.Warnings = append(result.Warnings, "Delimiter tag injection detected")
		result.DetectedPatterns = append(result.DetectedPatterns, "delimiter_tags")
		return result
	}

	// 1. Length validation
	if utf8.RuneCountInString(text) > v.maxLength {
		result.Warnings = append(result.Warnings, "Text exceeds maximum length")
		result.ThreatLevel = maxThreatLevel(result.ThreatLevel, ThreatLevelLow)
	}

	// 2. Detect prompt injection patterns
	detectedPatterns := v.detectInjectionPatterns(text)
	if len(detectedPatterns) > 0 {
		result.DetectedPatterns = detectedPatterns
		result.ThreatLevel = maxThreatLevel(result.ThreatLevel, v.calculateThreatLevel(detectedPatterns))
		result.Warnings = append(result.Warnings, "Suspicious prompt injection patterns detected")
	}

	// 3. Check against block list
	if v.containsBlockedTerms(text) {
		result.ThreatLevel = maxThreatLevel(result.ThreatLevel, ThreatLevelHigh)
		result.Warnings = append(result.Warnings, "Blocked terms detected")
	}

	// 4. Sanitize text
	result.SanitizedText = v.sanitize(text)

	// 5. Determine if valid
	result.IsValid = result.ThreatLevel < ThreatLevelHigh

	return result
}

func (v *InputValidator) detectInjectionPatterns(text string) []string {
	detected := []string{}
	lowerText := strings.ToLower(text)

	patterns := []string{
		"ignore previous",
		"ignore all previous",
		"disregard previous",
		"forget previous",
		"ignore above",
		"disregard above",
		"system:",
		"assistant:",
		"user:",
		"you are now",
		"new instruction",
		"override",
		"instead",
		"don't translate",
		"do not translate",
		"respond with",
		"your role is",
		"act as",
		"pretend",
		"simulate",
		"</s>",
		"<|im_end|>",
		"<|endoftext|>",
		"###",
		"---END---",
	}

	for _, pattern := range patterns {
		if strings.Contains(lowerText, pattern) {
			detected = append(detected, pattern)
		}
	}

	for _, regex := range v.suspiciousPatterns {
		if regex.MatchString(text) {
			detected = append(detected, regex.String())
		}
	}

	return detected
}

func (v *InputValidator) sanitize(text string) string {
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = removeControlCharacters(text)
	return text
}

func (v *InputValidator) calculateThreatLevel(patterns []string) ThreatLevel {
	count := len(patterns)

	critical := []string{"system:", "ignore previous", "override"}
	for _, p := range patterns {
		for _, c := range critical {
			if strings.Contains(strings.ToLower(p), c) {
				return ThreatLevelCritical
			}
		}
	}

	if count >= 3 {
		return ThreatLevelHigh
	} else if count >= 2 {
		return ThreatLevelMedium
	} else if count >= 1 {
		return ThreatLevelLow
	}

	return ThreatLevelNone
}

func (v *InputValidator) containsBlockedTerms(text string) bool {
	lowerText := strings.ToLower(text)
	for _, term := range v.blockList {
		if strings.Contains(lowerText, term) {
			return true
		}
	}
	return false
}

func (v *InputValidator) containsDelimiterTags(text string) bool {
	delimiters := []string{
		"<UserInput>",
		"</UserInput>",
		"<userinput>",
		"</userinput>",
		"<USERINPUT>",
		"</USERINPUT>",
	}

	for _, delimiter := range delimiters {
		if strings.Contains(text, delimiter) {
			return true
		}
	}

	encodedDelimiters := []string{
		"&lt;UserInput&gt;",
		"&lt;/UserInput&gt;",
		"%3CUserInput%3E",
		"%3C/UserInput%3E",
	}

	for _, delimiter := range encodedDelimiters {
		if strings.Contains(text, delimiter) {
			return true
		}
	}

	return false
}

func compileSuspiciousPatterns() []*regexp.Regexp {
	patterns := []string{
		`(?i)(you\s+are|you're)\s+(now|a|an)\s+\w+`,
		`(?i)(new|updated?)\s+(instruction|rule|command|prompt)`,
		`(?i)(system|assistant|user)\s*[:\-\=]`,
		"```",
		`[#\-=*]{3,}`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}

	return compiled
}

func loadBlockList() []string {
	return []string{}
}

func removeControlCharacters(text string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r == '\r' {
			return r
		}
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, text)
}

func maxThreatLevel(a, b ThreatLevel) ThreatLevel {
	if a > b {
		return a
	}
	return b
}

func (t ThreatLevel) String() string {
	switch t {
	case ThreatLevelNone:
		return "NONE"
	case ThreatLevelLow:
		return "LOW"
	case ThreatLevelMedium:
		return "MEDIUM"
	case ThreatLevelHigh:
		return "HIGH"
	case ThreatLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}
