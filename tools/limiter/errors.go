package limiter

import "errors"

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidKey        = errors.New("invalid rate limit key")
	ErrLimiterClosed     = errors.New("limiter is closed")
)
