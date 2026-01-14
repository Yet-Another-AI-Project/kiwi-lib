package middleware

import (
	"strings"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/utils"
	"github.com/gin-gonic/gin"
)

func NewBearerTokenAuth(token string) func(*gin.Context) {

	return func(c *gin.Context) {

		authorizationHeader := c.GetHeader("Authorization")

		if strings.HasPrefix(authorizationHeader, "Bearer ") {
			tokenInHeader := strings.TrimPrefix(authorizationHeader, "Bearer ")
			if tokenInHeader != token {
				utils.ResponseError(c, facade.ErrUnauthorized)
				return
			}
		} else {
			utils.ResponseError(c, facade.ErrUnauthorized)
			return
		}

		c.Next()
	}
}
