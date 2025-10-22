package translator

type Translator interface {
	Translate(text, sourceLanguage, targetLanguage string) (string, error)
	DetectLanguage(text string) (string, error)
}
