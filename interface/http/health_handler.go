package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type HealthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]CheckStatus `json:"checks"`
}

type CheckStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HealthCheckHandler struct {
	db     *sql.DB
	redis  *redis.Client
	logger *zap.Logger
}

func NewHealthCheckHandler(db *sql.DB, redis *redis.Client, logger *zap.Logger) *HealthCheckHandler {
	return &HealthCheckHandler{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

func (h *HealthCheckHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckStatus)

	// Check database
	dbStatus := h.checkDatabase(ctx)
	checks["database"] = dbStatus

	// Check Redis
	redisStatus := h.checkRedis(ctx)
	checks["redis"] = redisStatus

	// Determine overall status
	overallStatus := "ok"
	for _, check := range checks {
		if check.Status != "ok" {
			overallStatus = "unhealthy"
			break
		}
	}

	response := HealthResponse{
		Status: overallStatus,
		Checks: checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if overallStatus != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode health response", zap.Error(err))
	}
}

func (h *HealthCheckHandler) checkDatabase(ctx context.Context) CheckStatus {
	if h.db == nil {
		return CheckStatus{Status: "fail", Error: "database not initialized"}
	}

	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Error("Database health check failed", zap.Error(err))
		return CheckStatus{Status: "fail", Error: err.Error()}
	}

	return CheckStatus{Status: "ok"}
}

func (h *HealthCheckHandler) checkRedis(ctx context.Context) CheckStatus {
	if h.redis == nil {
		return CheckStatus{Status: "fail", Error: "redis not initialized"}
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		h.logger.Error("Redis health check failed", zap.Error(err))
		return CheckStatus{Status: "fail", Error: err.Error()}
	}

	return CheckStatus{Status: "ok"}
}

// HandleHealthGin handles health check requests with Gin framework
func (h *HealthCheckHandler) HandleHealthGin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckStatus)

	// Check database
	dbStatus := h.checkDatabase(ctx)
	checks["database"] = dbStatus

	// Check Redis
	redisStatus := h.checkRedis(ctx)
	checks["redis"] = redisStatus

	// Determine overall status
	overallStatus := "ok"
	for _, check := range checks {
		if check.Status != "ok" {
			overallStatus = "unhealthy"
			break
		}
	}

	response := HealthResponse{
		Status: overallStatus,
		Checks: checks,
	}

	if overallStatus != "ok" {
		c.JSON(http.StatusServiceUnavailable, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}
