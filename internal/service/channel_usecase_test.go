package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ntttrang/python-genai-your-slack-assistant/internal/model"
)

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

func TestCreateChannelConfig(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	config := &model.ChannelConfig{
		ChannelID:       "C123456",
		AutoTranslate:   true,
		SourceLanguages: []string{"en"},
		TargetLanguage:  "vi",
		Enabled:         true,
	}

	mockRepo.On("Save", config).Return(nil)
	mockCache.On("Delete", "channel_config:C123456").Return(nil)

	cu := NewChannelUseCase(mockRepo, mockCache)
	err := cu.CreateChannelConfig(config)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Save", config)
	mockCache.AssertCalled(t, "Delete", "channel_config:C123456")
}

func TestCreateChannelConfig_Error(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	config := &model.ChannelConfig{
		ChannelID: "C123456",
	}

	mockRepo.On("Save", config).Return(errors.New("database error"))

	cu := NewChannelUseCase(mockRepo, mockCache)
	err := cu.CreateChannelConfig(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create channel config")
}

func TestGetChannelConfig(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	config := &model.ChannelConfig{
		ID:              "1",
		ChannelID:       "C123456",
		AutoTranslate:   true,
		SourceLanguages: []string{"en"},
		TargetLanguage:  "vi",
		Enabled:         true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockCache.On("Get", "channel_config:C123456").Return("", errors.New("not found"))
	mockRepo.On("GetByChannelID", "C123456").Return(config, nil)
	mockCache.On("Set", "channel_config:C123456", "1", int64(3600)).Return(nil)

	cu := NewChannelUseCase(mockRepo, mockCache)
	result, err := cu.GetChannelConfig("C123456")

	assert.NoError(t, err)
	assert.Equal(t, config, result)
	mockRepo.AssertCalled(t, "GetByChannelID", "C123456")
}

func TestUpdateChannelConfig(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	config := &model.ChannelConfig{
		ID:              "1",
		ChannelID:       "C123456",
		AutoTranslate:   false,
		SourceLanguages: []string{"en", "fr"},
		TargetLanguage:  "vi",
		Enabled:         true,
	}

	mockRepo.On("Update", config).Return(nil)
	mockCache.On("Delete", "channel_config:C123456").Return(nil)

	cu := NewChannelUseCase(mockRepo, mockCache)
	err := cu.UpdateChannelConfig(config)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Update", config)
	mockCache.AssertCalled(t, "Delete", "channel_config:C123456")
}

func TestDeleteChannelConfig(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	mockRepo.On("Delete", "C123456").Return(nil)
	mockCache.On("Delete", "channel_config:C123456").Return(nil)

	cu := NewChannelUseCase(mockRepo, mockCache)
	err := cu.DeleteChannelConfig("C123456")

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Delete", "C123456")
	mockCache.AssertCalled(t, "Delete", "channel_config:C123456")
}

func TestListAllChannelConfigs(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	configs := []*model.ChannelConfig{
		{
			ID:        "1",
			ChannelID: "C123456",
			Enabled:   true,
		},
		{
			ID:        "2",
			ChannelID: "C789012",
			Enabled:   true,
		},
	}

	mockRepo.On("GetAll").Return(configs, nil)

	cu := NewChannelUseCase(mockRepo, mockCache)
	result, err := cu.ListAllChannelConfigs()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	mockRepo.AssertCalled(t, "GetAll")
}

func TestIsChannelEnabled(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	config := &model.ChannelConfig{
		ID:        "1",
		ChannelID: "C123456",
		Enabled:   true,
	}

	mockCache.On("Get", "channel_config:C123456").Return("", errors.New("not found"))
	mockRepo.On("GetByChannelID", "C123456").Return(config, nil)
	mockCache.On("Set", "channel_config:C123456", "1", int64(3600)).Return(nil)

	cu := NewChannelUseCase(mockRepo, mockCache)
	enabled, err := cu.IsChannelEnabled("C123456")

	assert.NoError(t, err)
	assert.True(t, enabled)
}

func TestIsChannelEnabled_NotFound(t *testing.T) {
	mockRepo := new(MockChannelRepository)
	mockCache := new(MockCache)

	mockCache.On("Get", "channel_config:C123456").Return("", errors.New("not found"))
	mockRepo.On("GetByChannelID", "C123456").Return(nil, errors.New("not found"))

	cu := NewChannelUseCase(mockRepo, mockCache)
	enabled, err := cu.IsChannelEnabled("C123456")

	// Default to enabled if config doesn't exist
	assert.NoError(t, err)
	assert.True(t, enabled)
}
