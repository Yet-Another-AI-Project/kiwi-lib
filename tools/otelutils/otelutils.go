package otelutils

import (
	"context"

	"github.com/gin-gonic/gin"

	oteltrace "go.opentelemetry.io/otel/trace"
)

func GetTraceID(ctx context.Context) string {

	// gin 特殊
	if c, ok := ctx.(*gin.Context); ok {
		spanContext := oteltrace.SpanContextFromContext(c.Request.Context())
		if spanContext.HasTraceID() {
			return spanContext.TraceID().String()
		}
		// return spanctx, span
	} else {
		spanContext := oteltrace.SpanContextFromContext(ctx)
		if spanContext.HasTraceID() {
			return spanContext.TraceID().String()
		}
	}

	return ""
}
