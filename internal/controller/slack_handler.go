package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ntttrang/go-genai-slack-assistant/internal/service/slack"
	"go.uber.org/zap"
)

type SlackWebhookHandler struct {
	eventProcessor slack.EventProcessor
	logger         *zap.Logger
}

func NewSlackWebhookHandler(eventProcessor slack.EventProcessor, logger *zap.Logger) *SlackWebhookHandler {
	return &SlackWebhookHandler{
		eventProcessor: eventProcessor,
		logger:         logger,
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

	// Process event
	go h.eventProcessor.ProcessEvent(r.Context(), payload)

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

	// Process event
	go h.eventProcessor.ProcessEvent(c.Request.Context(), payload)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
