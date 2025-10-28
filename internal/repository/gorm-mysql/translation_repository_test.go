package gormmysql

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTranslationRepositoryImpl_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTranslationRepository(db)

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

	mock.ExpectExec("INSERT INTO translations").
		WithArgs(
			translation.ID,
			translation.SourceMessageID,
			translation.SourceText,
			translation.SourceLanguage,
			translation.TargetLanguage,
			translation.TranslatedText,
			translation.Hash,
			translation.UserID,
			translation.ChannelID,
			sqlmock.AnyArg(),
			translation.TTL,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(translation)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepositoryImpl_GetByHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTranslationRepository(db)
	hash := "abc123"

	rows := sqlmock.NewRows([]string{
		"id", "source_message_id", "source_text", "source_language",
		"target_language", "translated_text", "hash", "user_id",
		"channel_id", "created_at", "ttl",
	}).AddRow(
		"test-id-1", "msg-123", "Hello", "English",
		"Vietnamese", "Xin chào", hash, "user-1",
		"channel-1", time.Now(), int64(3600),
	)

	mock.ExpectQuery("SELECT (.+) FROM translations WHERE hash").
		WithArgs(hash).
		WillReturnRows(rows)

	result, err := repo.GetByHash(hash)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Hello", result.SourceText)
	assert.Equal(t, "Xin chào", result.TranslatedText)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepositoryImpl_GetByHash_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTranslationRepository(db)
	hash := "nonexistent"

	mock.ExpectQuery("SELECT (.+) FROM translations WHERE hash").
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "source_message_id", "source_text", "source_language",
			"target_language", "translated_text", "hash", "user_id",
			"channel_id", "created_at", "ttl",
		}))

	result, err := repo.GetByHash(hash)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepositoryImpl_GetByChannelID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTranslationRepository(db)
	channelID := "channel-1"
	limit := 10

	rows := sqlmock.NewRows([]string{
		"id", "source_message_id", "source_text", "source_language",
		"target_language", "translated_text", "hash", "user_id",
		"channel_id", "created_at", "ttl",
	}).AddRow(
		"test-id-1", "msg-123", "Hello", "English",
		"Vietnamese", "Xin chào", "hash1", "user-1",
		channelID, time.Now(), int64(3600),
	).AddRow(
		"test-id-2", "msg-456", "Goodbye", "English",
		"Vietnamese", "Tạm biệt", "hash2", "user-1",
		channelID, time.Now(), int64(3600),
	)

	mock.ExpectQuery("SELECT (.+) FROM translations WHERE channel_id").
		WithArgs(channelID, limit).
		WillReturnRows(rows)

	results, err := repo.GetByChannelID(channelID, limit)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Hello", results[0].SourceText)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepositoryImpl_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTranslationRepository(db)
	id := "test-id-1"

	rows := sqlmock.NewRows([]string{
		"id", "source_message_id", "source_text", "source_language",
		"target_language", "translated_text", "hash", "user_id",
		"channel_id", "created_at", "ttl",
	}).AddRow(
		id, "msg-123", "Hello", "English",
		"Vietnamese", "Xin chào", "abc123", "user-1",
		"channel-1", time.Now(), int64(3600),
	)

	mock.ExpectQuery("SELECT (.+) FROM translations WHERE id").
		WithArgs(id).
		WillReturnRows(rows)

	result, err := repo.GetByID(id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "Hello", result.SourceText)
	assert.NoError(t, mock.ExpectationsWereMet())
}
