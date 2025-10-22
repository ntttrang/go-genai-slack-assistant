package model

type Translator interface {
	Translate(text, sourceLanguage, targetLanguage string) (string, error)
	DetectLanguage(text string) (string, error)
}

type TranslationRequest struct {
	Text           string
	SourceLanguage string
	TargetLanguage string
}

type TranslationResponse struct {
	OriginalText   string
	TranslatedText string
	SourceLanguage string
	TargetLanguage string
}
