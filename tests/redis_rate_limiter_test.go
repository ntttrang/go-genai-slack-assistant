package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ntttrang/go-genai-slack-assistant/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisRateLimiter_CheckUserLimit_FirstRequest(t *testing.T) {
	// Skip if Redis not available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	// Cleanup
	ctx := context.Background()
	client.FlushDB(ctx)

	limiter := ratelimit.NewRedisRateLimiter(client)

	allowed, remaining, _, err := limiter.CheckUserLimit("user123")

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, ratelimit.UserRateLimit, remaining)
}

func TestRedisRateLimiter_IncrementUserLimit(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	ctx := context.Background()
	client.FlushDB(ctx)

	limiter := ratelimit.NewRedisRateLimiter(client)

	err := limiter.IncrementUserLimit("user123")
	assert.NoError(t, err)

	allowed, remaining, _, err := limiter.CheckUserLimit("user123")
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, ratelimit.UserRateLimit-1, remaining)
}

func TestRedisRateLimiter_CheckChannelLimit_FirstRequest(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	ctx := context.Background()
	client.FlushDB(ctx)

	limiter := ratelimit.NewRedisRateLimiter(client)

	allowed, remaining, _, err := limiter.CheckChannelLimit("channel123")

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, ratelimit.ChannelRateLimit, remaining)
}

func TestRedisRateLimiter_RateLimitExceeded(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	ctx := context.Background()
	client.FlushDB(ctx)

	limiter := ratelimit.NewRedisRateLimiter(client)

	// Increment to limit
	for i := 0; i < ratelimit.UserRateLimit; i++ {
		err := limiter.IncrementUserLimit("user123")
		require.NoError(t, err)
	}

	// Check should fail
	allowed, remaining, _, err := limiter.CheckUserLimit("user123")

	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
}

func TestRedisRateLimiter_TTLExpiration(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	ctx := context.Background()
	client.FlushDB(ctx)

	limiter := ratelimit.NewRedisRateLimiter(client)

	// Set a very short TTL for testing
	key := "rate_limit:user:testuser"
	client.Set(ctx, key, "10", time.Second)

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	allowed, _, _, err := limiter.CheckUserLimit("testuser")

	assert.NoError(t, err)
	assert.True(t, allowed)
}
