package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewMetricsHandler(t *testing.T) {
	metricsManager := metrics.NewMetrics()
	logger, _ := zap.NewProduction()

	handler := NewMetricsHandler(metricsManager, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.metrics)
	assert.NotNil(t, handler.logger)
}

func TestMetricsHandler_HandleMetricsGin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	metricsManager := metrics.NewMetrics()
	logger, _ := zap.NewProduction()

	handler := NewMetricsHandler(metricsManager, logger)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/metrics", nil)

	handler.HandleMetricsGin(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, response)
}

func TestMetricsHandler_HandleMetrics_StandardHTTP(t *testing.T) {
	metricsManager := metrics.NewMetrics()
	logger, _ := zap.NewProduction()

	handler := NewMetricsHandler(metricsManager, logger)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.HandleMetrics(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response is valid JSON
	assert.NotNil(t, response)
}

func TestMetricsHandler_WithMetricsData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	metricsManager := metrics.NewMetrics()
	logger, _ := zap.NewProduction()

	// Record some metrics
	metricsManager.RecordTranslationRequest("user123", "channel456", 100, true)

	handler := NewMetricsHandler(metricsManager, logger)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/metrics", nil)

	handler.HandleMetricsGin(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify we get back some data
	assert.NotNil(t, response)
}
