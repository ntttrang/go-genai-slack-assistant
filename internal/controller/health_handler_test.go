package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewHealthCheckHandler(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	redisClient, _ := redismock.NewClientMock()
	logger, _ := zap.NewProduction()

	handler := NewHealthCheckHandler(db, redisClient, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.db)
	assert.NotNil(t, handler.redis)
	assert.NotNil(t, handler.logger)
}

func TestHealthCheckHandler_HandleHealthGin_AllHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	redisClient, redisMock := redismock.NewClientMock()
	logger, _ := zap.NewProduction()

	handler := NewHealthCheckHandler(db, redisClient, logger)

	// Expect successful DB ping
	mock.ExpectPing()

	// Expect successful Redis ping
	redisMock.ExpectPing().SetVal("PONG")

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/health", nil)

	handler.HandleHealthGin(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.Contains(t, rec.Body.String(), `"database":{"status":"ok"}`)
	assert.Contains(t, rec.Body.String(), `"redis":{"status":"ok"}`)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestHealthCheckHandler_HandleHealthGin_DatabaseUnhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	redisClient, redisMock := redismock.NewClientMock()
	logger, _ := zap.NewProduction()

	handler := NewHealthCheckHandler(db, redisClient, logger)

	// Expect DB ping to fail
	mock.ExpectPing().WillReturnError(errors.New("database connection failed"))

	// Expect successful Redis ping
	redisMock.ExpectPing().SetVal("PONG")

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/health", nil)

	handler.HandleHealthGin(ctx)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"unhealthy"`)
	assert.Contains(t, rec.Body.String(), `"database":{"status":"fail"`)
	assert.Contains(t, rec.Body.String(), "database connection failed")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestHealthCheckHandler_HandleHealthGin_RedisUnhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	redisClient, redisMock := redismock.NewClientMock()
	logger, _ := zap.NewProduction()

	handler := NewHealthCheckHandler(db, redisClient, logger)

	// Expect successful DB ping
	mock.ExpectPing()

	// Expect Redis ping to fail
	redisMock.ExpectPing().SetErr(errors.New("redis connection failed"))

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/health", nil)

	handler.HandleHealthGin(ctx)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"unhealthy"`)
	assert.Contains(t, rec.Body.String(), `"redis":{"status":"fail"`)
	assert.Contains(t, rec.Body.String(), "redis connection failed")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestHealthCheckHandler_HandleHealthGin_NilDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	redisClient, redisMock := redismock.NewClientMock()
	logger, _ := zap.NewProduction()

	handler := NewHealthCheckHandler(nil, redisClient, logger)

	// Expect successful Redis ping
	redisMock.ExpectPing().SetVal("PONG")

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/health", nil)

	handler.HandleHealthGin(ctx)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"unhealthy"`)
	assert.Contains(t, rec.Body.String(), "database not initialized")

	// Verify all expectations were met
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestHealthCheckHandler_HandleHealthGin_NilRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	logger, _ := zap.NewProduction()

	handler := NewHealthCheckHandler(db, nil, logger)

	// Expect successful DB ping
	mock.ExpectPing()

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/health", nil)

	handler.HandleHealthGin(ctx)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"unhealthy"`)
	assert.Contains(t, rec.Body.String(), "redis not initialized")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthCheckHandler_CheckDatabase(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	logger, _ := zap.NewProduction()
	handler := NewHealthCheckHandler(db, nil, logger)

	t.Run("successful ping", func(t *testing.T) {
		mock.ExpectPing()
		status := handler.checkDatabase(context.Background())
		assert.Equal(t, "ok", status.Status)
		assert.Empty(t, status.Error)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("failed ping", func(t *testing.T) {
		mock.ExpectPing().WillReturnError(sql.ErrConnDone)
		status := handler.checkDatabase(context.Background())
		assert.Equal(t, "fail", status.Status)
		assert.Contains(t, status.Error, "conn")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("nil database", func(t *testing.T) {
		handlerNilDB := NewHealthCheckHandler(nil, nil, logger)
		status := handlerNilDB.checkDatabase(context.Background())
		assert.Equal(t, "fail", status.Status)
		assert.Equal(t, "database not initialized", status.Error)
	})
}

func TestHealthCheckHandler_CheckRedis(t *testing.T) {
	redisClient, redisMock := redismock.NewClientMock()
	logger, _ := zap.NewProduction()
	handler := NewHealthCheckHandler(nil, redisClient, logger)

	t.Run("successful ping", func(t *testing.T) {
		redisMock.ExpectPing().SetVal("PONG")
		status := handler.checkRedis(context.Background())
		assert.Equal(t, "ok", status.Status)
		assert.Empty(t, status.Error)
		assert.NoError(t, redisMock.ExpectationsWereMet())
	})

	t.Run("failed ping", func(t *testing.T) {
		redisMock.ExpectPing().SetErr(errors.New("connection refused"))
		status := handler.checkRedis(context.Background())
		assert.Equal(t, "fail", status.Status)
		assert.Contains(t, status.Error, "connection refused")
		assert.NoError(t, redisMock.ExpectationsWereMet())
	})

	t.Run("nil redis", func(t *testing.T) {
		handlerNilRedis := NewHealthCheckHandler(nil, nil, logger)
		status := handlerNilRedis.checkRedis(context.Background())
		assert.Equal(t, "fail", status.Status)
		assert.Equal(t, "redis not initialized", status.Error)
	})
}

func TestHealthCheckHandler_HandleHealth_StandardHTTP(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	redisClient, redisMock := redismock.NewClientMock()
	logger, _ := zap.NewProduction()

	handler := NewHealthCheckHandler(db, redisClient, logger)

	// Expect successful DB ping
	mock.ExpectPing()

	// Expect successful Redis ping
	redisMock.ExpectPing().SetVal("PONG")

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	handler.HandleHealth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}
