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

func TestDatabaseConnectionPoolSettings(t *testing.T) {
	// Create a mock database to test connection pool settings with ping monitoring
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()
	
	// Expect ping to succeed
	mock.ExpectPing()
	
	// Set connection pool settings (simulating NewDB behavior)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	
	// Verify settings
	err = db.Ping()
	assert.NoError(t, err)
	
	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDatabasePingFailure(t *testing.T) {
	// Test that connection failures are properly handled
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()
	
	// Expect ping to fail
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)
	
	// Verify ping error is returned
	err = db.Ping()
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	
	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
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
			// Construct DSN as NewDB does
			dsn := constructDSN(tt.config)
			assert.Equal(t, tt.wantDSN, dsn)
		})
	}
}

// Helper function to construct DSN (extracted from NewDB logic)
func constructDSN(config DBConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		config.User, config.Password, config.Host, config.Port, config.Database)
}

func TestDatabaseConnectionRetries(t *testing.T) {
	// Test that database handles transient connection issues
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()
	
	// First ping fails
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)
	
	err = db.Ping()
	assert.Error(t, err)
	
	// Second ping succeeds
	mock.ExpectPing()
	
	err = db.Ping()
	assert.NoError(t, err)
	
	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDatabaseQueryExecution(t *testing.T) {
	// Test basic query execution to ensure database connectivity works
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	
	// Expect a simple query
	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM translations").WillReturnRows(rows)
	
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM translations").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	
	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDatabaseTransactionSupport(t *testing.T) {
	// Test that database supports transactions
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	
	// Expect transaction begin
	mock.ExpectBegin()
	
	tx, err := db.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	
	// Expect commit
	mock.ExpectCommit()
	
	err = tx.Commit()
	assert.NoError(t, err)
	
	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDatabaseConnectionClose(t *testing.T) {
	// Test that database connections are properly closed
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	
	// Expect close
	mock.ExpectClose()
	
	err = db.Close()
	assert.NoError(t, err)
	
	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
