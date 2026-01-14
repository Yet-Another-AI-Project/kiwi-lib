package limiter

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Redis 限流器
type RedisLimiter struct {
	client *redis.Client
	key    string // Redis 存储的 key
	max    int
}

func NewRedisLimiter(client *redis.Client, key string, max int) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		key:    key,
		max:    max,
	}
}

func (r *RedisLimiter) Acquire() (bool, error) {
	// 使用 Lua 保证原子性
	script := `
        local current = redis.call('GET', KEYS[1]) or 0
        if tonumber(current) >= tonumber(ARGV[1]) then
            return 0
        else
            redis.call('INCR', KEYS[1])
            return 1
        end
    `
	result, err := r.client.Eval(context.Background(), script, []string{r.key}, r.max).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

func (r *RedisLimiter) Release() error {
	return r.client.Decr(context.Background(), r.key).Err()
}

func (r *RedisLimiter) Count() (int, error) {
	val, err := r.client.Get(context.Background(), r.key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}
