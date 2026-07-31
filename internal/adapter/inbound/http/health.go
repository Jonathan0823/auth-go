package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DBPinger interface {
	Ping(context.Context) error
}

type HealthHandler struct {
	db DBPinger
}

const readinessTimeout = 2 * time.Second

func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Live godoc
// @Summary Check process liveness
// @ID healthLive
// @Description Returns liveness without checking external dependencies.
// @Tags operations
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready godoc
// @Summary Check service readiness
// @ID healthReady
// @Description Checks PostgreSQL connectivity with a bounded timeout.
// @Tags operations
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
