package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	UserRateLimit    = 10  // 10 translations per minute
	ChannelRateLimit = 30  // 30 translations per minute
	RateLimitWindow  = 60  // 1 minute in seconds
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (r *RedisRateLimiter) CheckUserLimit(userID string) (bool, int, int64, error) {
	key := fmt.Sprintf("rate_limit:user:%s", userID)
	return r.checkLimit(key, UserRateLimit)
}

func (r *RedisRateLimiter) CheckChannelLimit(channelID string) (bool, int, int64, error) {
	key := fmt.Sprintf("rate_limit:channel:%s", channelID)
	return r.checkLimit(key, ChannelRateLimit)
}

func (r *RedisRateLimiter) IncrementUserLimit(userID string) error {
	key := fmt.Sprintf("rate_limit:user:%s", userID)
	return r.increment(key)
}

func (r *RedisRateLimiter) IncrementChannelLimit(channelID string) error {
	key := fmt.Sprintf("rate_limit:channel:%s", channelID)
	return r.increment(key)
}

func (r *RedisRateLimiter) checkLimit(key string, limit int) (bool, int, int64, error) {
	ctx := context.Background()

	count, err := r.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, 0, 0, fmt.Errorf("failed to check rate limit: %w", err)
	}

	allowed := count < limit
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return false, 0, 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	var resetTime int64
	if ttl == -1 || ttl == -2 {
		resetTime = time.Now().Add(time.Duration(RateLimitWindow) * time.Second).Unix()
	} else {
		resetTime = time.Now().Add(ttl).Unix()
	}

	return allowed, remaining, resetTime, nil
}

func (r *RedisRateLimiter) increment(key string) error {
	ctx := context.Background()

	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(RateLimitWindow)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return nil
}
