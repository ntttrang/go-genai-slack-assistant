package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ntttrang/go-genai-slack-assistant/pkg/metrics"
	"go.uber.org/zap"
)

type MetricsHandler struct {
	metrics *metrics.Metrics
	logger  *zap.Logger
}

func NewMetricsHandler(metrics *metrics.Metrics, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		metrics: metrics,
		logger:  logger,
	}
}

func (h *MetricsHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	stats := h.metrics.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("Failed to encode metrics response", zap.Error(err))
	}
}

// HandleMetricsGin handles metrics requests with Gin framework
func (h *MetricsHandler) HandleMetricsGin(c *gin.Context) {
	stats := h.metrics.GetStats()
	c.JSON(http.StatusOK, stats)
}
