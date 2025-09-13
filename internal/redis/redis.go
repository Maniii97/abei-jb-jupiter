package redisconn

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(url string) *RedisClient {
	opt, _ := redis.ParseURL(url)

	// Optimized connection pool for high-traffic scenarios
	opt.PoolSize = 50 // this increases the number of connections in the pool, better for high concurrency
	opt.MinIdleConns = 10
	opt.MaxIdleConns = 25
	opt.ConnMaxLifetime = 30 * time.Minute
	opt.ConnMaxIdleTime = 5 * time.Minute
	opt.MaxRetries = 3
	opt.MaxRetryBackoff = 512 * time.Millisecond

	client := redis.NewClient(opt)
	return &RedisClient{Client: client}
}

func (r *RedisClient) LockSeat(ctx context.Context, seatID string, ttl time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, "lock:"+seatID, "locked", ttl).Result()
}

func (r *RedisClient) UnlockSeat(ctx context.Context, seatID string) error {
	return r.Client.Del(ctx, "lock:"+seatID).Err()
}
