package utils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func CtxWithTimeOut(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 10*time.Second)
}
