package gormmysql

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTranslationRepositoryImpl_Save(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	sqlDB, _ := gormDB.DB()
	defer closeMockDB(t, sqlDB, mock)
	repo := NewTranslationRepository(gormDB)

	translation := &model.Translation{
		ID:              "test-id-1",
		SourceMessageID: "msg-123",
		SourceText:      "Hello",
		SourceLanguage:  "English",
		TargetLanguage:  "Vietnamese",
		TranslatedText:  "Xin chào",
		Hash:            "abc123",
		UserID:          "user-1",
		ChannelID:       "channel-1",
		CreatedAt:       time.Now(),
		TTL:             3600,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `translations`").
		WithArgs(translation.ID, translation.SourceMessageID, translation.SourceText, translation.SourceLanguage, translation.TargetLanguage, translation.TranslatedText, translation.Hash, translation.UserID, translation.ChannelID, sqlmock.AnyArg(), translation.TTL).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Save(translation)

	assert.NoError(t, err)
}

func TestTranslationRepositoryImpl_GetByHash(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		hash           string
		mockSetup      func(sqlmock.Sqlmock, string, time.Time)
		validateResult func(*testing.T, *model.Translation, error)
	}{
		{
			name: "found translation",
			hash: "abc123",
			mockSetup: func(mock sqlmock.Sqlmock, hash string, now time.Time) {
				rows := sqlmock.NewRows([]string{"id", "source_message_id", "source_text", "source_language", "target_language", "translated_text", "hash", "user_id", "channel_id", "created_at", "ttl"}).
					AddRow("test-id-1", "msg-123", "Hello", "English", "Vietnamese", "Xin chào", hash, "user-1", "channel-1", now, 3600)
				mock.ExpectQuery("SELECT \\* FROM `translations` WHERE hash = \\?").
					WithArgs(hash, 1).
					WillReturnRows(rows)
			},
			validateResult: func(t *testing.T, result *model.Translation, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Hello", result.SourceText)
				assert.Equal(t, "Xin chào", result.TranslatedText)
			},
		},
		{
			name: "translation not found",
			hash: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock, hash string, now time.Time) {
				rows := sqlmock.NewRows([]string{"id", "source_message_id", "source_text", "source_language", "target_language", "translated_text", "hash", "user_id", "channel_id", "created_at", "ttl"})
				mock.ExpectQuery("SELECT \\* FROM `translations` WHERE hash = \\?").
					WithArgs(hash, 1).
					WillReturnRows(rows)
			},
			validateResult: func(t *testing.T, result *model.Translation, err error) {
				assert.NoError(t, err)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock := setupMockDB(t)
			sqlDB, _ := gormDB.DB()
			defer closeMockDB(t, sqlDB, mock)
			repo := NewTranslationRepository(gormDB)

			tt.mockSetup(mock, tt.hash, now)

			result, err := repo.GetByHash(tt.hash)

			tt.validateResult(t, result, err)
		})
	}
}

func TestTranslationRepositoryImpl_GetByChannelID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	sqlDB, _ := gormDB.DB()
	defer closeMockDB(t, sqlDB, mock)
	repo := NewTranslationRepository(gormDB)
	channelID := "channel-1"
	limit := 10
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "source_message_id", "source_text", "source_language", "target_language", "translated_text", "hash", "user_id", "channel_id", "created_at", "ttl"}).
		AddRow("test-id-1", "msg-123", "Hello", "English", "Vietnamese", "Xin chào", "hash1", "user-1", channelID, now, 3600).
		AddRow("test-id-2", "msg-456", "Goodbye", "English", "Vietnamese", "Tạm biệt", "hash2", "user-1", channelID, now, 3600)

	mock.ExpectQuery("SELECT \\* FROM `translations` WHERE channel_id = \\? ORDER BY created_at DESC LIMIT \\?").
		WithArgs(channelID, limit).
		WillReturnRows(rows)

	results, err := repo.GetByChannelID(channelID, limit)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Hello", results[0].SourceText)
}

func TestTranslationRepositoryImpl_GetByID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	sqlDB, _ := gormDB.DB()
	defer closeMockDB(t, sqlDB, mock)
	repo := NewTranslationRepository(gormDB)
	id := "test-id-1"
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "source_message_id", "source_text", "source_language", "target_language", "translated_text", "hash", "user_id", "channel_id", "created_at", "ttl"}).
		AddRow(id, "msg-123", "Hello", "English", "Vietnamese", "Xin chào", "abc123", "user-1", "channel-1", now, 3600)

	mock.ExpectQuery("SELECT \\* FROM `translations` WHERE id = \\?").
		WithArgs(id, 1).
		WillReturnRows(rows)

	result, err := repo.GetByID(id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "Hello", result.SourceText)
}
