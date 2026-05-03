package controllers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func requestContext(c *gin.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), timeout)
}
