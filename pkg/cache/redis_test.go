package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache_SetAndGet(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	err = cache.Set("test-key", "test-value", 3600)
	assert.NoError(t, err)

	val, err := cache.Get("test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", val)
}

func TestRedisCache_GetNonExistent(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	_, err = cache.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestRedisCache_Delete(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	_ = cache.Set("test-key", "test-value", 3600)
	err = cache.Delete("test-key")
	assert.NoError(t, err)

	_, err = cache.Get("test-key")
	assert.Error(t, err)
}

func TestRedisCache_Exists_True(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	_ = cache.Set("test-key", "test-value", 3600)
	exists, err := cache.Exists("test-key")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisCache_Exists_False(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	exists, err := cache.Exists("nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisCache_TTL(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	err = cache.Set("test-key", "test-value", 1)
	assert.NoError(t, err)

	exists, err := cache.Exists("test-key")
	assert.NoError(t, err)
	assert.True(t, exists)

	mr.FastForward(2 * time.Second)

	_, err = cache.Get("test-key")
	assert.Error(t, err)
}

func TestRedisCache_MultipleOperations(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	_ = cache.Set("key1", "value1", 3600)
	_ = cache.Set("key2", "value2", 3600)
	_ = cache.Set("key3", "value3", 3600)

	val1, _ := cache.Get("key1")
	val2, _ := cache.Get("key2")
	val3, _ := cache.Get("key3")

	assert.Equal(t, "value1", val1)
	assert.Equal(t, "value2", val2)
	assert.Equal(t, "value3", val3)

	_ = cache.Delete("key2")

	_, err = cache.Get("key2")
	assert.Error(t, err)

	exists1, _ := cache.Exists("key1")
	exists3, _ := cache.Exists("key3")
	assert.True(t, exists1)
	assert.True(t, exists3)
}

func TestRedisCache_UpdateValue(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	_ = cache.Set("test-key", "initial-value", 3600)
	val, _ := cache.Get("test-key")
	assert.Equal(t, "initial-value", val)

	_ = cache.Set("test-key", "updated-value", 3600)
	val, _ = cache.Get("test-key")
	assert.Equal(t, "updated-value", val)
}

func TestRedisCache_EmptyValue(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	err = cache.Set("test-key", "", 3600)
	assert.NoError(t, err)

	val, err := cache.Get("test-key")
	assert.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestRedisCache_LongValue(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache("127.0.0.1", mr.Server().Addr().Port, "")
	assert.NoError(t, err)

	longValue := ""
	for i := 0; i < 1000; i++ {
		longValue += "x"
	}

	err = cache.Set("test-key", longValue, 3600)
	assert.NoError(t, err)

	val, err := cache.Get("test-key")
	assert.NoError(t, err)
	assert.Equal(t, longValue, val)
}
