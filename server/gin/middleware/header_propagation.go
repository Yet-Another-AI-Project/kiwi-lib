package middleware

import (
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/header"
	"github.com/gin-gonic/gin"
)

// HeaderPropagationMiddleware 头部透传中间件
func HeaderPropagationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := header.ExtractHeadersToContext(c.Request.Context(), c.Request)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
