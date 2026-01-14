package middleware

import (
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/utils"
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/limiter"
	"github.com/gin-gonic/gin"
)

func NewLimiter(limiter limiter.Limiter) func(*gin.Context) {

	return func(c *gin.Context) {
		ok, err := limiter.Acquire()
		if err != nil {
			utils.ResponseError(c, facade.ErrServerInternal.Wrap(err))
			return
		}

		if !ok {
			utils.ResponseError(c, facade.ErrTooManyRequests)
			return
		}

		defer limiter.Release()

		c.Next()
	}
}
