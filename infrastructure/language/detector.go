package language

import (
	"fmt"

	"github.com/pemistahl/lingua-go"
)

type LanguageDetector struct {
	detector lingua.LanguageDetector
}

func NewLanguageDetector() (*LanguageDetector, error) {
	languages := []lingua.Language{
		lingua.English,
		lingua.Vietnamese,
		lingua.Spanish,
		lingua.French,
		lingua.German,
		lingua.Chinese,
		lingua.Japanese,
		lingua.Korean,
	}

	detector, err := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to create language detector: %w", err)
	}

	return &LanguageDetector{detector: detector}, nil
}

func (ld *LanguageDetector) DetectLanguage(text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("empty text provided")
	}

	lang, exists := ld.detector.DetectLanguageOf(text)
	if !exists {
		return "", fmt.Errorf("unable to detect language")
	}

	return lang.String(), nil
}

func (ld *LanguageDetector) GetLanguageCode(langStr string) (string, error) {
	codeMap := map[string]string{
		"ENGLISH":    "en",
		"VIETNAMESE": "vi",
		"SPANISH":    "es",
		"FRENCH":     "fr",
		"GERMAN":     "de",
		"CHINESE":    "zh",
		"JAPANESE":   "ja",
		"KOREAN":     "ko",
	}

	code, exists := codeMap[langStr]
	if !exists {
		return "", fmt.Errorf("unsupported language: %s", langStr)
	}

	return code, nil
}
