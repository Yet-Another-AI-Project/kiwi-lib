package limiter

import (
	"sync/atomic"
)

// 内存限流器
type MemoryLimiter struct {
	max     int   // 最大连接数
	current int64 // 当前连接数（原子操作）
}

func NewMemoryLimiter(maxConnections int) *MemoryLimiter {
	limiter := &MemoryLimiter{max: maxConnections}

	return limiter
}

func (m *MemoryLimiter) Acquire() (bool, error) {
	current := atomic.LoadInt64(&m.current)
	if current >= int64(m.max) {
		return false, nil
	}
	atomic.AddInt64(&m.current, 1)
	return true, nil
}

func (m *MemoryLimiter) Release() error {
	atomic.AddInt64(&m.current, -1)
	return nil
}

func (m *MemoryLimiter) Count() (int64, error) {
	return atomic.LoadInt64(&m.current), nil
}

func (m *MemoryLimiter) Rate() (int64, error) {
	panic("implement me")
}
