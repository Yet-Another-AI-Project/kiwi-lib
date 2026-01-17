package futurxgin

import (
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/gin/middleware"
	"github.com/futurxlab/golanggraph/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	ginlogger = func(logger logger.ILogger) gin.HandlerFunc {
		return func(c *gin.Context) {
			start := time.Now()
			// some evil middlewares modify this values
			path := c.Request.URL.Path

			// inject logger to context
			c.Set("logger", logger)

			c.Next()

			// traceid
			traceID := c.Writer.Header().Get("x-trace-id")

			// log request
			end := time.Now()
			latency := end.Sub(start)

			logger.Infof(c, "Gin Request Log status: %d, path: %s, method: %s, ip: %s, useragent: %s, latency: %s, trace_id: %s",
				c.Writer.Status(),
				path,
				c.Request.Method,
				c.ClientIP(),
				c.Request.UserAgent(),
				latency,
				traceID,
			)
		}
	}
	metricsEndpoint string
	serviceName     string
)

type Option func(*gin.Engine) error

func WithServiceName(name string) Option {
	return func(engine *gin.Engine) error {
		serviceName = name
		return nil
	}
}

func WithMetricsEndpoint(ep string) Option {
	return func(engine *gin.Engine) error {
		metricsEndpoint = ep
		return nil
	}
}

func WithLogger(logger logger.ILogger) Option {
	return func(engine *gin.Engine) error {
		engine.Use(ginlogger(logger))
		return nil
	}
}

func NewGin(options ...Option) (*gin.Engine, error) {
	engine := gin.New()
	for _, option := range options {
		if err := option(engine); err != nil {
			return nil, err
		}
	}
	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"},
		AllowHeaders: []string{"*"},
		MaxAge:       12 * time.Hour,
	}))

	if serviceName != "" {
		handler, err := middleware.OpenTelemetryMiddleware(serviceName, metricsEndpoint)
		if err != nil {
			return nil, err
		}
		engine.Use(handler)
		engine.Use(middleware.ReponseTraceID())
		engine.Use(middleware.MetricsMiddleware(serviceName))
	}

	engine.ContextWithFallback = true

	return engine, nil
}
