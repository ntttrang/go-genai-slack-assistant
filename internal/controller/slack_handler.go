package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ntttrang/go-genai-slack-assistant/internal/model"
	"github.com/ntttrang/go-genai-slack-assistant/internal/queue"
	"go.uber.org/zap"
)

type SlackWebhookHandler struct {
	workerPool *queue.WorkerPool
	logger     *zap.Logger
	seqCounter uint64
}

func NewSlackWebhookHandler(workerPool *queue.WorkerPool, logger *zap.Logger) *SlackWebhookHandler {
	return &SlackWebhookHandler{
		workerPool: workerPool,
		logger:     logger,
	}
}

func (h *SlackWebhookHandler) HandleSlackEvents(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("Failed to unmarshal payload", zap.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Handle URL verification challenge
	if eventType, ok := payload["type"].(string); ok && eventType == "url_verification" {
		challenge, ok := payload["challenge"].(string)
		if !ok {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	// Extract message event and enqueue for processing
	event, err := h.extractMessageEvent(payload)
	if err != nil {
		h.logger.Debug("Skipping non-message event or unable to extract event details", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
		return
	}

	// Enqueue event for ordered processing
	h.workerPool.Enqueue(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

func (h *SlackWebhookHandler) HandleSlackEventsGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer c.Request.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("Failed to unmarshal payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	h.logger.Info("Received Slack event", zap.String("type", fmt.Sprintf("%v", payload["type"])))

	// Handle URL verification challenge
	if eventType, ok := payload["type"].(string); ok && eventType == "url_verification" {
		challenge, ok := payload["challenge"].(string)
		if !ok {
			h.logger.Error("Challenge parameter missing or invalid")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}
		h.logger.Info("Responding to URL verification challenge", zap.String("challenge", challenge))
		c.Data(http.StatusOK, "text/plain", []byte(challenge))
		return
	}

	// Extract message event and enqueue for processing
	event, err := h.extractMessageEvent(payload)
	if err != nil {
		h.logger.Debug("Skipping non-message event or unable to extract event details", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// Enqueue event for ordered processing
	h.workerPool.Enqueue(event)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// extractMessageEvent extracts relevant fields from the Slack payload and creates a MessageEvent.
func (h *SlackWebhookHandler) extractMessageEvent(payload map[string]interface{}) (*model.MessageEvent, error) {
	// Get event_id if available
	eventID, _ := payload["event_id"].(string)

	// Get event callback data
	eventCallback, ok := payload["event"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing event callback in payload")
	}

	// Extract channel_id, user_id, and message timestamp
	channelID, _ := eventCallback["channel"].(string)
	userID, _ := eventCallback["user"].(string)
	messageTS, _ := eventCallback["ts"].(string)

	// For reaction events, extract from item
	if channelID == "" || messageTS == "" {
		if item, ok := eventCallback["item"].(map[string]interface{}); ok {
			if channelID == "" {
				channelID, _ = item["channel"].(string)
			}
			if messageTS == "" {
				messageTS, _ = item["ts"].(string)
			}
		}
	}

	// Validate we have minimum required fields
	if channelID == "" {
		return nil, fmt.Errorf("missing channel_id in event")
	}

	// User ID might be empty for some event types, use a default
	if userID == "" {
		userID = "unknown"
	}

	return &model.MessageEvent{
		EventID:    eventID,
		ChannelID:  channelID,
		UserID:     userID,
		MessageTS:  messageTS,
		Payload:    payload,
		ReceivedAt: time.Now(),
		Sequence:   atomic.AddUint64(&h.seqCounter, 1),
	}, nil
}
