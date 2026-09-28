package http

import (
	"context"
	"net/http"
	"time"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
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
// @Success 200 {object} dto.HealthResponse
// @Router /health/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}

// Ready godoc
// @Summary Check service readiness
// @ID healthReady
// @Description Checks PostgreSQL connectivity with a bounded timeout.
// @Tags operations
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 503 {object} dto.HealthResponse
// @Router /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.HealthResponse{Status: "not_ready"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.HealthResponse{Status: "not_ready"})
		return
	}
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "ready"})
}
