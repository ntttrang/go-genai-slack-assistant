package testutils

import (
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockTranslationRepository mocks the TranslationRepository interface
type MockTranslationRepository struct {
	mock.Mock
}

func (m *MockTranslationRepository) Save(translation *model.Translation) error {
	args := m.Called(translation)
	return args.Error(0)
}

func (m *MockTranslationRepository) GetByHash(hash string) (*model.Translation, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Translation), args.Error(1)
}

func (m *MockTranslationRepository) GetByID(id string) (*model.Translation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Translation), args.Error(1)
}

func (m *MockTranslationRepository) GetByChannelID(channelID string, limit int) ([]*model.Translation, error) {
	args := m.Called(channelID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Translation), args.Error(1)
}

// MockChannelRepository mocks the ChannelRepository interface
type MockChannelRepository struct {
	mock.Mock
}

func (m *MockChannelRepository) Save(config *model.ChannelConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockChannelRepository) GetByChannelID(channelID string) (*model.ChannelConfig, error) {
	args := m.Called(channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ChannelConfig), args.Error(1)
}

func (m *MockChannelRepository) Update(config *model.ChannelConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockChannelRepository) Delete(channelID string) error {
	args := m.Called(channelID)
	return args.Error(0)
}

func (m *MockChannelRepository) GetAll() ([]*model.ChannelConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ChannelConfig), args.Error(1)
}

// MockCache mocks the Cache interface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(key string, value string, ttl int64) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) Exists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

// MockTranslator mocks the Translator interface
type MockTranslator struct {
	mock.Mock
}

func (m *MockTranslator) Translate(text, sourceLang, targetLang string) (string, error) {
	args := m.Called(text, sourceLang, targetLang)
	return args.String(0), args.Error(1)
}

func (m *MockTranslator) DetectLanguage(text string) (string, error) {
	args := m.Called(text)
	return args.String(0), args.Error(1)
}

func (m *MockTranslator) Close() error {
	args := m.Called()
	return args.Error(0)
}

