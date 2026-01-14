package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metrics struct {
	requestCount metric.Int64Counter
	responseTime metric.Float64Histogram
}

func MetricsMiddleware(serviceName string) gin.HandlerFunc {
	meter := otel.Meter("gin-middleware")
	metrics := initMeter(meter)

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds() * 1000

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		labels := []attribute.KeyValue{
			attribute.String("route", route),
			attribute.String("method", c.Request.Method),
			attribute.Int("status", c.Writer.Status()),
			attribute.String("service", serviceName),
		}

		metrics.requestCount.Add(c.Request.Context(), 1, metric.WithAttributeSet(attribute.NewSet(labels...)))
		metrics.responseTime.Record(c.Request.Context(), duration, metric.WithAttributeSet(attribute.NewSet(labels...)))
	}
}

func initMeter(meter metric.Meter) *metrics {
	requestCount, err := meter.Int64Counter(
		"http.request.count",
		metric.WithDescription("Total number of requests"),
	)
	if err != nil {
		panic(err)
	}

	responseTime, err := meter.Float64Histogram(
		"http.response.time.ms",
		metric.WithDescription("Response time in ms"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(0, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000, 10000, 30000, 60000, 120000),
	)
	if err != nil {
		panic(err)
	}

	return &metrics{
		requestCount: requestCount,
		responseTime: responseTime,
	}
}
