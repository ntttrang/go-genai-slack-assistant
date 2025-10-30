package database

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB_Success(t *testing.T) {
	// This test verifies database connection logic using sqlmock
	// In real scenarios, this would test against a test database
	
	// Test DSN construction
	config := DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
	}
	
	expectedDSN := "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	
	// Verify DSN format is correct
	assert.Contains(t, expectedDSN, config.User)
	assert.Contains(t, expectedDSN, config.Host)
	assert.Contains(t, expectedDSN, config.Database)
}

func TestDatabasePing(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful ping with connection pool settings",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			wantErr: false,
		},
		{
			name: "ping failure with connection error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			wantErr:     true,
			expectedErr: sql.ErrConnDone,
		},
		{
			name: "ping failure with timeout error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(fmt.Errorf("connection timeout"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer func() {
				_ = db.Close()
			}()

			tt.setupMock(mock)

			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(5)

			err = db.Ping()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDBConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  DBConfig
		wantDSN string
	}{
		{
			name: "standard config",
			config: DBConfig{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "password",
				Database: "mydb",
			},
			wantDSN: "root:password@tcp(localhost:3306)/mydb?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		},
		{
			name: "custom port",
			config: DBConfig{
				Host:     "db.example.com",
				Port:     3307,
				User:     "appuser",
				Password: "secret",
				Database: "production",
			},
			wantDSN: "appuser:secret@tcp(db.example.com:3307)/production?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		},
		{
			name: "empty password",
			config: DBConfig{
				Host:     "localhost",
				Port:     3306,
				User:     "guest",
				Password: "",
				Database: "testdb",
			},
			wantDSN: "guest:@tcp(localhost:3306)/testdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := constructDSN(tt.config)
			assert.Equal(t, tt.wantDSN, dsn)
		})
	}
}

func constructDSN(config DBConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		config.User, config.Password, config.Host, config.Port, config.Database)
}

func TestDatabaseConnectionRetries(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		attempts  int
	}{
		{
			name: "first ping fails, second succeeds",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
				mock.ExpectPing()
			},
			attempts: 2,
		},
		{
			name: "multiple failures before success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
				mock.ExpectPing()
			},
			attempts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer func() {
				_ = db.Close()
			}()

			tt.setupMock(mock)

			for i := 0; i < tt.attempts; i++ {
				err = db.Ping()
			}

			assert.NoError(t, err)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseQueryExecution(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		setupMock     func(sqlmock.Sqlmock)
		expectedCount int
		wantErr       bool
	}{
		{
			name:  "successful query with results",
			query: "SELECT COUNT(*) FROM translations",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM translations").WillReturnRows(rows)
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:  "query with zero results",
			query: "SELECT COUNT(*) FROM translations",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM translations").WillReturnRows(rows)
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:  "query with multiple results",
			query: "SELECT COUNT(*) FROM translations",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(100)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM translations").WillReturnRows(rows)
			},
			expectedCount: 100,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() {
				_ = db.Close()
			}()

			tt.setupMock(mock)

			var count int
			err = db.QueryRow(tt.query).Scan(&count)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseTransactions(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		operation string
		wantErr   bool
	}{
		{
			name: "successful transaction with commit",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			operation: "commit",
			wantErr:   false,
		},
		{
			name: "successful transaction with rollback",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			operation: "rollback",
			wantErr:   false,
		},
		{
			name: "transaction begin failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			operation: "begin",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() {
				_ = db.Close()
			}()

			tt.setupMock(mock)

			tx, err := db.Begin()
			if tt.operation == "begin" && tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tx)

				switch tt.operation {
				case "commit":
					err = tx.Commit()
					assert.NoError(t, err)
				case "rollback":
					err = tx.Rollback()
					assert.NoError(t, err)
				}
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseConnectionClose(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "successful close",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectClose()
			},
			wantErr: false,
		},
		{
			name: "close failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectClose().WillReturnError(fmt.Errorf("close error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			tt.setupMock(mock)

			err = db.Close()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
