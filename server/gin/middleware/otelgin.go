package middleware

import (
	"context"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/otelutils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func OpenTelemetryMiddleware(serviceName string, endpoint string) (gin.HandlerFunc, error) {
	if err := setupOTelSDK(serviceName, endpoint); err != nil {
		return nil, err
	}

	return otelgin.Middleware(serviceName), nil
}

func ReponseTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := otelutils.GetTraceID(c)

		if traceID != "" {
			c.Writer.Header().Set("x-trace-id", traceID)
		}

		c.Next()
	}
}

func setupOTelSDK(serviceName string, endpoint string) error {

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName), // 设置服务名称
		),
	)

	if err != nil {
		return nil
	}

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	if endpoint != "" {
		meterProvider, err := newMeterProvider(res, endpoint)
		if err != nil {
			return err
		}
		otel.SetMeterProvider(meterProvider)
	}

	// 如果不设置 TraceProvider，无法生成TraceID
	tracerProvider, err := newTraceProvider(res)
	if err != nil {
		return err
	}
	otel.SetTracerProvider(tracerProvider)

	return nil
}

func newMeterProvider(res *resource.Resource, endpoint string) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpoint(endpoint), // Prometheus 服务器的地址和端口
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithURLPath("/api/v1/otlp/v1/metrics"), // OTLP 接收路径
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(
			metricExporter,
			metric.WithInterval(15*time.Second))),
	)

	return meterProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	traceProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	return traceProvider, nil
}
