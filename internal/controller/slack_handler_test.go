package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/ntttrang/go-genai-slack-assistant/internal/queue"
	"github.com/ntttrang/go-genai-slack-assistant/internal/testutils/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSlackWebhookHandlerURLVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventProc := mocks.NewMockEventProcessorService(ctrl)
	logger, _ := zap.NewProduction()
	
	// Create worker pool with mock processor
	workerPool := queue.NewWorkerPool(mockEventProc, 10, 1*time.Minute, logger)
	defer workerPool.Shutdown(5 * time.Second)

	handler := NewSlackWebhookHandler(workerPool, logger)

	// Create request body with URL verification challenge
	payload := map[string]interface{}{
		"type":      "url_verification",
		"challenge": "test-challenge-123",
	}
	body, _ := json.Marshal(payload)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("POST", "/slack/events", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.HandleSlackEventsGin(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test-challenge-123", rec.Body.String())
}

func TestSlackWebhookHandlerEventCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventProc := mocks.NewMockEventProcessorService(ctrl)
	logger, _ := zap.NewProduction()
	
	// Create worker pool with mock processor
	workerPool := queue.NewWorkerPool(mockEventProc, 10, 1*time.Minute, logger)
	defer workerPool.Shutdown(5 * time.Second)

	handler := NewSlackWebhookHandler(workerPool, logger)

	// Create request body with regular event callback
	payload := map[string]interface{}{
		"type": "event_callback",
		"event": map[string]interface{}{
			"type":    "message",
			"channel": "C123",
			"user":    "U456",
			"ts":      "1234567890.123456",
		},
	}
	body, _ := json.Marshal(payload)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("POST", "/slack/events", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Setup expectation for ProcessEvent call (will be called by worker)
	mockEventProc.EXPECT().
		ProcessEvent(gomock.Any(), gomock.Any()).
		Times(1).
		Do(func(ctx interface{}, payload interface{}) {
			// Simulate processing
		})

	// Execute
	handler.HandleSlackEventsGin(ctx)

	// Assert - response should be OK (event is enqueued for processing)
	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Wait for worker to process
	time.Sleep(100 * time.Millisecond)
}

func TestSlackWebhookHandlerInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventProc := mocks.NewMockEventProcessorService(ctrl)
	logger, _ := zap.NewProduction()
	
	// Create worker pool with mock processor
	workerPool := queue.NewWorkerPool(mockEventProc, 10, 1*time.Minute, logger)
	defer workerPool.Shutdown(5 * time.Second)

	handler := NewSlackWebhookHandler(workerPool, logger)

	// Create request with invalid JSON
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("POST", "/slack/events", bytes.NewBufferString("invalid json"))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.HandleSlackEventsGin(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSlackWebhookHandlerImplementsCorrectSignature(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventProc := mocks.NewMockEventProcessorService(ctrl)
	logger, _ := zap.NewProduction()
	
	// Create worker pool with mock processor
	workerPool := queue.NewWorkerPool(mockEventProc, 10, 1*time.Minute, logger)
	defer workerPool.Shutdown(5 * time.Second)

	handler := NewSlackWebhookHandler(workerPool, logger)

	// Verify the handler is created correctly
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.workerPool)
}
