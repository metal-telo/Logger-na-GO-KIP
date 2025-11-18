package handler

import (
	"context"
)

// Дополнительные middleware можно добавить здесь
type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
)

// Пример дополнительного middleware
func (h *Handler) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}