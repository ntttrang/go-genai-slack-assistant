package testutils

//go:generate mockgen -destination=mocks/mock_translation_repository.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/service TranslationRepository
//go:generate mockgen -destination=mocks/mock_channel_repository.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/service ChannelRepository
//go:generate mockgen -destination=mocks/mock_cache.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/service Cache
//go:generate mockgen -destination=mocks/mock_translation_service.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/service TranslationService
//go:generate mockgen -destination=mocks/mock_channel_service.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/service ChannelService
//go:generate mockgen -destination=mocks/mock_event_processor_service.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/service EventProcessorService
//go:generate mockgen -destination=mocks/mock_event_processor.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/service/slack EventProcessor
//go:generate mockgen -destination=mocks/mock_translator.go -package=mocks github.com/ntttrang/go-genai-slack-assistant/internal/translator Translator
