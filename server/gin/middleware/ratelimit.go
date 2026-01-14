package middleware

import (
	"errors"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/utils"
	"github.com/futurxlab/golanggraph/logger"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/limiter"

	"github.com/gin-gonic/gin"
)

type LimitType string

const (
	// LimitTypeUser 用户限流
	LimitTypeUser LimitType = "user"
	// LimitTypeAPI 接口限流
	LimitTypeAPI LimitType = "api"
)

// LimiterConfig 限流器配置
type LimiterConfig struct {
	Type        LimitType // 限流类型
	WindowSize  int64     // 窗口大小（秒）
	MaxRequests int64     // 窗口内最大请求数
}

// RateLimit 创建限流中间件
func RateLimit(config LimiterConfig, logger logger.ILogger) func(*gin.Context) {
	// 创建限流器
	fixedWindowLimiter := limiter.NewFixedWindowLimiter(time.Duration(config.WindowSize)*time.Second, config.MaxRequests, logger)

	return func(c *gin.Context) {
		var key string

		// 根据限流类型获取key
		switch config.Type {
		case LimitTypeUser:
			// 从上下文获取用户ID
			userInfo, exists := c.Get("user_id")
			if !exists {
				utils.ResponseError(c, facade.ErrTooManyRequests.Facade("未获取到user id 信息"))
				return
			}
			key = userInfo.(string) // 假设userInfo是用户ID字符串

		case LimitTypeAPI:
			// 使用接口路径作为key
			key = c.FullPath()
			if key == "" {
				key = c.Request.URL.Path
			}

		default:
			c.Next()
			return
		}

		// 判断是否允许请求通过
		allowed, err := fixedWindowLimiter.Allow(c, key)
		if err != nil {
			if errors.Is(err, limiter.ErrLimiterClosed) {
				c.Next()
				return
			}
			utils.ResponseError(c, facade.ErrTooManyRequests.Facade("操作太频繁，休息一下"))
			return
		}

		if !allowed {
			utils.ResponseError(c, facade.ErrTooManyRequests.Facade("操作太频繁，休息一下"))
			return
		}

		// 请求通过，继续处理
		c.Next()
	}
}
