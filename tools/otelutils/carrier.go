package otelutils

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/propagation"
)

func MapCarrier(ctx context.Context) map[string]string {
	// 6. 向后传递 Header: traceparent
	pp := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	)

	carrier := propagation.MapCarrier{}

	if c, ok := ctx.(*gin.Context); ok {
		pp.Inject(c.Request.Context(), carrier)
	} else {
		pp.Inject(ctx, carrier)
	}

	return carrier
}
