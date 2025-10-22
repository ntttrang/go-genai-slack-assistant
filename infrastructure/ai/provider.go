package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(apiKey string, model string) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

func (gp *GeminiProvider) Translate(text, sourceLanguage, targetLanguage string) (string, error) {
	ctx := context.Background()

	prompt := fmt.Sprintf(`Translate the following text from %s to %s. Provide only the translated text without any additional explanation or formatting:

Text: "%s"`, sourceLanguage, targetLanguage, text)

	model := gp.client.GenerativeModel(gp.model)
	model.Temperature = 0.1
	model.TopP = 0.9

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate translation: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format from Gemini")
	}

	return string(textPart), nil
}

func (gp *GeminiProvider) Close() error {
	return gp.client.Close()
}
