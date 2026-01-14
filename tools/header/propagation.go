package header

import (
	"context"
	"net/http"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/otelutils"
)

// 业务相关需要透传的 HTTP 头部常量
const (
	HeaderXRequestID      = "X-Request-Id"
	HeaderXTraceID        = "X-Trace-Id"
	HeaderXUserID         = "X-User-Id"
	HeaderXOrgID          = "X-Org-Id"
	HeaderXSessionID      = "X-Session-Id"
	HeaderXTenantID       = "X-Tenant-Id"
	HeaderXClientID       = "X-Client-Id"
	HeaderXAppVersion     = "X-App-Version"
	HeaderXDeviceID       = "X-Device-Id"
	HeaderXPlatform       = "X-Platform"
	HeaderXRealIP         = "X-Real-IP"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXForwardedProto = "X-Forwarded-Proto"
	HeaderXForwardedHost  = "X-Forwarded-Host"
	HeaderAuthorization   = "Authorization"
	HeaderUserAgent       = "User-Agent"
	HeaderContentLanguage = "Content-Language"
)

// PropagatedHeaders 需要透传的HTTP头部列表（业务头部）
var PropagatedHeaders = []string{
	HeaderXRequestID,
	HeaderXTraceID,
	HeaderXUserID,
	HeaderXOrgID,
	HeaderXSessionID,
	HeaderXTenantID,
	HeaderXClientID,
	HeaderXAppVersion,
	HeaderXDeviceID,
	HeaderXPlatform,
	HeaderXRealIP,
	HeaderXForwardedFor,
	HeaderXForwardedProto,
	HeaderXForwardedHost,
	HeaderAuthorization,
	HeaderUserAgent,
	HeaderContentLanguage,
}

// contextKeyHeaders 用于在 context 中存储头部的 key
type contextKeyHeaders struct{}

// ExtractHeadersToContext 从HTTP请求中提取头部并存入context
func ExtractHeadersToContext(ctx context.Context, r *http.Request) context.Context {
	headers := make(map[string]string)

	// 1. 保留 OpenTelemetry trace context（traceparent 等）
	if traceHeaders := otelutils.MapCarrier(ctx); traceHeaders != nil {
		for k, v := range traceHeaders {
			headers[k] = v
		}
	}

	// 2. 提取业务头部
	for _, key := range PropagatedHeaders {
		if val := r.Header.Get(key); val != "" {
			headers[key] = val
		}
	}

	return context.WithValue(ctx, contextKeyHeaders{}, headers)
}

// ApplyPropagatedHeaders 从 context 中获取透传的头部并应用到 http.Header
// 直接修改传入的 header，无需调用者手动遍历
func ApplyPropagatedHeaders(ctx context.Context, header http.Header) {
	if headers, ok := ctx.Value(contextKeyHeaders{}).(map[string]string); ok {
		for key, value := range headers {
			header.Set(key, value)
		}
	}
}

// GetPropagatedHeader 根据常量 key 获取在上下文中保存的对应头部值
// 未找到时返回空字符串
func GetPropagatedHeader(ctx context.Context, key string) string {
	if headers, ok := ctx.Value(contextKeyHeaders{}).(map[string]string); ok {
		if v, exists := headers[key]; exists {
			return v
		}
	}
	return ""
}
