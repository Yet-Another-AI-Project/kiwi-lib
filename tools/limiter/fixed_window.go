package limiter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/futurxlab/golanggraph/logger"
)

// FixedWindowLimiter 固定窗口限流器
type FixedWindowLimiter struct {
	windowSize  time.Duration
	maxRequests int64
	limiters    sync.Map // key -> *windowLimiter
	closed      int32
	logger      logger.ILogger
}

// windowLimiter 单个窗口限流器
type windowLimiter struct {
	currentCount int64
	windowStart  time.Time
	mu           sync.Mutex
}

// NewFixedWindowLimiter 创建固定窗口限流器
func NewFixedWindowLimiter(windowSize time.Duration, maxRequests int64, logger logger.ILogger) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		logger:      logger,
	}
}

// Allow 判断是否允许请求通过
func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if atomic.LoadInt32(&l.closed) == 1 {
		l.logger.Infof(ctx, "fixed window limiter is closed, key: %s", key)
		return false, ErrLimiterClosed
	}

	if key == "" {
		l.logger.Infof(ctx, "fixed window limiter key is empty")
		return false, ErrInvalidKey
	}

	// 获取或创建窗口限流器
	limiter, _ := l.limiters.LoadOrStore(key, &windowLimiter{
		windowStart: time.Now(),
	})

	wl := limiter.(*windowLimiter)
	return wl.allow(l.windowSize, l.maxRequests)
}

// Close 关闭限流器
func (l *FixedWindowLimiter) Close() error {
	atomic.StoreInt32(&l.closed, 1)
	return nil
}

// allow 判断单个窗口是否允许请求通过
func (wl *windowLimiter) allow(windowSize time.Duration, maxRequests int64) (bool, error) {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	now := time.Now()
	if now.Sub(wl.windowStart) >= windowSize {
		// 重置窗口
		wl.windowStart = now
		wl.currentCount = 0
	}

	if wl.currentCount >= maxRequests {
		return false, ErrRateLimitExceeded
	}

	wl.currentCount++
	return true, nil
}
