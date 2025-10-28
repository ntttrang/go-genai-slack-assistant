package gormmysql

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestChannelRepositoryImpl_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)

	config := &model.ChannelConfig{
		ChannelID:       "C123456",
		TargetLanguage:  "Vietnamese",
		Enabled:         true,
		SourceLanguages: []string{"English"},
	}

	mock.ExpectExec("INSERT INTO channel_configs").
		WithArgs(
			config.ChannelID,
			config.AutoTranslate,
			sqlmock.AnyArg(),
			config.TargetLanguage,
			config.Enabled,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(config)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_GetByChannelID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)
	channelID := "C123456"
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "channel_id", "auto_translate", "source_languages",
		"target_language", "enabled", "created_at", "updated_at",
	}).AddRow(
		"1", channelID, true, "English",
		"Vietnamese", true, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM channel_configs WHERE channel_id").
		WithArgs(channelID).
		WillReturnRows(rows)

	result, err := repo.GetByChannelID(channelID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, channelID, result.ChannelID)
	assert.Equal(t, "Vietnamese", result.TargetLanguage)
	assert.True(t, result.Enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_GetByChannelID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)
	channelID := "C999999"

	mock.ExpectQuery("SELECT (.+) FROM channel_configs WHERE channel_id").
		WithArgs(channelID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "channel_id", "auto_translate", "source_languages",
			"target_language", "enabled", "created_at", "updated_at",
		}))

	result, err := repo.GetByChannelID(channelID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)

	config := &model.ChannelConfig{
		ChannelID:       "C123456",
		TargetLanguage:  "English",
		Enabled:         false,
		SourceLanguages: []string{"Vietnamese"},
	}

	mock.ExpectExec("UPDATE channel_configs").
		WithArgs(
			config.AutoTranslate,
			sqlmock.AnyArg(),
			config.TargetLanguage,
			config.Enabled,
			config.ChannelID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(config)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)

	config := &model.ChannelConfig{
		ChannelID: "C999999",
	}

	mock.ExpectExec("UPDATE channel_configs").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(config)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)
	channelID := "C123456"

	mock.ExpectExec("DELETE FROM channel_configs WHERE channel_id").
		WithArgs(channelID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(channelID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)
	channelID := "C999999"

	mock.ExpectExec("DELETE FROM channel_configs WHERE channel_id").
		WithArgs(channelID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(channelID)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "channel_id", "auto_translate", "source_languages",
		"target_language", "enabled", "created_at", "updated_at",
	}).AddRow(
		"1", "C123456", true, "English",
		"Vietnamese", true, now, now,
	).AddRow(
		"2", "C789012", true, "French",
		"Spanish", true, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM channel_configs").
		WillReturnRows(rows)

	results, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "C123456", results[0].ChannelID)
	assert.Equal(t, "C789012", results[1].ChannelID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChannelRepositoryImpl_GetAll_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewChannelRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "channel_id", "auto_translate", "source_languages",
		"target_language", "enabled", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM channel_configs").
		WillReturnRows(rows)

	results, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Len(t, results, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
