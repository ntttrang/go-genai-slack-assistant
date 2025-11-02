package middleware

import (
	"fmt"

	"github.com/ntttrang/go-genai-slack-assistant/pkg/security"
	"go.uber.org/zap"
)

type SecurityMiddleware struct {
	inputValidator   *security.InputValidator
	outputValidator  *security.OutputValidator
	logger           *zap.Logger
	blockHighThreat  bool
	logSuspicious    bool
}

func NewSecurityMiddleware(
	inputValidator *security.InputValidator,
	outputValidator *security.OutputValidator,
	logger *zap.Logger,
	blockHighThreat bool,
	logSuspicious bool,
) *SecurityMiddleware {
	return &SecurityMiddleware{
		inputValidator:  inputValidator,
		outputValidator: outputValidator,
		logger:          logger,
		blockHighThreat: blockHighThreat,
		logSuspicious:   logSuspicious,
	}
}

func (sm *SecurityMiddleware) ValidateInput(text string) (security.ValidationResult, error) {
	result := sm.inputValidator.Validate(text)

	if result.ThreatLevel >= security.ThreatLevelMedium && sm.logSuspicious {
		sm.logger.Warn("Suspicious input detected",
			zap.String("text_preview", truncate(text, 50)),
			zap.String("threat_level", result.ThreatLevel.String()),
			zap.Strings("detected_patterns", result.DetectedPatterns),
			zap.Strings("warnings", result.Warnings))
	}

	if sm.blockHighThreat && result.ThreatLevel >= security.ThreatLevelHigh {
		sm.logger.Error("High threat input blocked",
			zap.String("text_preview", truncate(text, 50)),
			zap.String("threat_level", result.ThreatLevel.String()),
			zap.Strings("detected_patterns", result.DetectedPatterns))

		return result, fmt.Errorf("input blocked due to security concerns: %s", result.ThreatLevel.String())
	}

	return result, nil
}

func (sm *SecurityMiddleware) ValidateOutput(output, originalInput string) (security.OutputValidationResult, error) {
	result := sm.outputValidator.ValidateTranslation(output, originalInput)

	if !result.IsValid {
		sm.logger.Error("Invalid translation output",
			zap.String("output_preview", truncate(output, 50)),
			zap.Strings("issues", result.Issues))

		return result, fmt.Errorf("output validation failed: %v", result.Issues)
	}

	return result, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
