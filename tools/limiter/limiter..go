package limiter

type Limiter interface {
	// 尝试增加计数，返回是否允许连接
	Acquire() (bool, error)
	// 释放计数
	Release() error
	// 当前连接数
	Count() (int64, error)
	// 当前qps
	Rate() (int64, error)
}
