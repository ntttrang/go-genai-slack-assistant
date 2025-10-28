package gormmysql

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func closeMockDB(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	require.NoError(t, err, "not all database expectations were met")

	db.Close()
}

func TestChannelRepositoryImpl_Save(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		config      *model.ChannelConfig
		mockSetup   func(sqlmock.Sqlmock, *model.ChannelConfig)
		expectError bool
	}{
		{
			name: "successful save",
			config: &model.ChannelConfig{
				ID:              "test-1",
				ChannelID:       "C123456",
				TargetLanguage:  "Vietnamese",
				Enabled:         true,
				SourceLanguages: `["English"]`,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, config *model.ChannelConfig) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `channel_configs`").
					WithArgs(config.ID, config.ChannelID, config.AutoTranslate, config.SourceLanguages, config.TargetLanguage, config.Enabled, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "save with auto translate enabled",
			config: &model.ChannelConfig{
				ID:              "test-2",
				ChannelID:       "C789012",
				AutoTranslate:   true,
				TargetLanguage:  "Spanish",
				Enabled:         true,
				SourceLanguages: `["French"]`,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, config *model.ChannelConfig) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `channel_configs`").
					WithArgs(config.ID, config.ChannelID, config.AutoTranslate, config.SourceLanguages, config.TargetLanguage, config.Enabled, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock := setupMockDB(t)
			sqlDB, _ := gormDB.DB()
			defer closeMockDB(t, sqlDB, mock)
			repo := NewChannelRepository(gormDB)

			tt.mockSetup(mock, tt.config)

			err := repo.Save(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChannelRepositoryImpl_GetByChannelID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		channelID     string
		mockSetup     func(sqlmock.Sqlmock, string, time.Time)
		expectError   bool
		validateResult func(*testing.T, *model.ChannelConfig)
	}{
		{
			name:      "found channel config",
			channelID: "C123456",
			mockSetup: func(mock sqlmock.Sqlmock, channelID string, now time.Time) {
				rows := sqlmock.NewRows([]string{"id", "channel_id", "auto_translate", "source_languages", "target_language", "enabled", "created_at", "updated_at"}).
					AddRow("test-1", channelID, true, `["English"]`, "Vietnamese", true, now, now)
				mock.ExpectQuery("SELECT \\* FROM `channel_configs` WHERE channel_id = \\?").
					WithArgs(channelID, 1).
					WillReturnRows(rows)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *model.ChannelConfig) {
				assert.NotNil(t, result)
				assert.Equal(t, "C123456", result.ChannelID)
				assert.Equal(t, "Vietnamese", result.TargetLanguage)
				assert.True(t, result.Enabled)
			},
		},
		{
			name:      "channel not found",
			channelID: "C999999",
			mockSetup: func(mock sqlmock.Sqlmock, channelID string, now time.Time) {
				rows := sqlmock.NewRows([]string{"id", "channel_id", "auto_translate", "source_languages", "target_language", "enabled", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM `channel_configs` WHERE channel_id = \\?").
					WithArgs(channelID, 1).
					WillReturnRows(rows)
			},
			expectError: true,
			validateResult: func(t *testing.T, result *model.ChannelConfig) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock := setupMockDB(t)
			sqlDB, _ := gormDB.DB()
			defer closeMockDB(t, sqlDB, mock)
			repo := NewChannelRepository(gormDB)

			tt.mockSetup(mock, tt.channelID, now)

			result, err := repo.GetByChannelID(tt.channelID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.validateResult(t, result)
		})
	}
}

func TestChannelRepositoryImpl_Update(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		config      *model.ChannelConfig
		mockSetup   func(sqlmock.Sqlmock, *model.ChannelConfig)
		expectError bool
	}{
		{
			name: "successful update",
			config: &model.ChannelConfig{
				ID:              "test-1",
				ChannelID:       "C123456",
				TargetLanguage:  "English",
				Enabled:         false,
				SourceLanguages: `["Vietnamese"]`,
				UpdatedAt:       now,
			},
			mockSetup: func(mock sqlmock.Sqlmock, config *model.ChannelConfig) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `channel_configs` SET").
					WithArgs(config.AutoTranslate, config.Enabled, `["Vietnamese"]`, config.TargetLanguage, sqlmock.AnyArg(), config.ChannelID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "update not found",
			config: &model.ChannelConfig{
				ID:              "nonexistent",
				ChannelID:       "C999999",
				TargetLanguage:  "English",
				Enabled:         true,
				SourceLanguages: `["Vietnamese"]`,
			},
			mockSetup: func(mock sqlmock.Sqlmock, config *model.ChannelConfig) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `channel_configs` SET").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), config.ChannelID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock := setupMockDB(t)
			sqlDB, _ := gormDB.DB()
			defer closeMockDB(t, sqlDB, mock)
			repo := NewChannelRepository(gormDB)

			tt.mockSetup(mock, tt.config)

			err := repo.Update(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChannelRepositoryImpl_Delete(t *testing.T) {
	tests := []struct {
		name        string
		channelID   string
		mockSetup   func(sqlmock.Sqlmock, string)
		expectError bool
	}{
		{
			name:      "successful delete",
			channelID: "C123456",
			mockSetup: func(mock sqlmock.Sqlmock, channelID string) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `channel_configs` WHERE channel_id = \\?").
					WithArgs(channelID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:      "delete not found",
			channelID: "C999999",
			mockSetup: func(mock sqlmock.Sqlmock, channelID string) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `channel_configs` WHERE channel_id = \\?").
					WithArgs(channelID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock := setupMockDB(t)
			sqlDB, _ := gormDB.DB()
			defer closeMockDB(t, sqlDB, mock)
			repo := NewChannelRepository(gormDB)

			tt.mockSetup(mock, tt.channelID)

			err := repo.Delete(tt.channelID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChannelRepositoryImpl_GetAll(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		mockSetup     func(sqlmock.Sqlmock, time.Time)
		expectedCount int
		validateResults func(*testing.T, []*model.ChannelConfig)
	}{
		{
			name: "get multiple configs",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				rows := sqlmock.NewRows([]string{"id", "channel_id", "auto_translate", "source_languages", "target_language", "enabled", "created_at", "updated_at"}).
					AddRow("test-1", "C123456", true, `["English"]`, "Vietnamese", true, now, now).
					AddRow("test-2", "C789012", true, `["French"]`, "Spanish", true, now, now)
				mock.ExpectQuery("SELECT \\* FROM `channel_configs` ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedCount: 2,
			validateResults: func(t *testing.T, results []*model.ChannelConfig) {
				assert.Equal(t, "C123456", results[0].ChannelID)
				assert.Equal(t, "C789012", results[1].ChannelID)
			},
		},
		{
			name: "get empty list",
			mockSetup: func(mock sqlmock.Sqlmock, now time.Time) {
				rows := sqlmock.NewRows([]string{"id", "channel_id", "auto_translate", "source_languages", "target_language", "enabled", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT \\* FROM `channel_configs` ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedCount: 0,
			validateResults: func(t *testing.T, results []*model.ChannelConfig) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock := setupMockDB(t)
			sqlDB, _ := gormDB.DB()
			defer closeMockDB(t, sqlDB, mock)
			repo := NewChannelRepository(gormDB)

			tt.mockSetup(mock, now)

			results, err := repo.GetAll()

			assert.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)
			tt.validateResults(t, results)
		})
	}
}
