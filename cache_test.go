package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EwanGreer/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	ID   uint
	Name string
}

func (t *TestStruct) CacheKey() string    { return fmt.Sprintf("%d_%s", t.ID, t.Name) }
func (t *TestStruct) CachePrefix() string { return "prefix" }

const connection = "localhost:6379"

func TestCacheInit(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) { return nil, nil })
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCacheGet(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) { return nil, nil })
	assert.NoError(t, err)

	ts := &TestStruct{
		ID:   1,
		Name: "James",
	}

	err = c.Set(context.Background(), ts)
	assert.NoError(t, err)

	val, err := c.Get(context.Background(), ts.CacheKey())
	assert.NoError(t, err)
	assert.NotNil(t, val)
}

func TestCacheStore_CallBack(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return &TestStruct{
			ID:   2,
			Name: "Ryan",
		}, nil
	})
	assert.NoError(t, err)

	result, err := c.Get(context.Background(), "nothing_here")
	assert.NoError(t, err)

	assert.Equal(t, uint(2), result.ID)
	assert.Equal(t, "Ryan", result.Name)
}

func TestCacheDelete(t *testing.T) {
	callbackCalled := false
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		callbackCalled = true
		return &TestStruct{
			ID:   3,
			Name: "Sarah",
		}, nil
	})
	assert.NoError(t, err)

	ts := &TestStruct{
		ID:   1,
		Name: "James",
	}

	err = c.Set(context.Background(), ts)
	assert.NoError(t, err)

	val, err := c.Get(context.Background(), ts.CacheKey())
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, uint(1), val.ID)
	assert.Equal(t, "James", val.Name)
	assert.False(t, callbackCalled)

	err = c.Delete(context.Background(), ts.CacheKey())
	assert.NoError(t, err)

	val, err = c.Get(context.Background(), ts.CacheKey())
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, uint(3), val.ID)
	assert.Equal(t, "Sarah", val.Name)
	assert.True(t, callbackCalled)
}

func TestCacheDelete_Multiple(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return nil, fmt.Errorf("should not be called")
	})
	assert.NoError(t, err)

	ts1 := &TestStruct{
		ID:   1,
		Name: "James",
	}
	ts2 := &TestStruct{
		ID:   2,
		Name: "Ryan",
	}

	err = c.Set(context.Background(), ts1)
	assert.NoError(t, err)
	err = c.Set(context.Background(), ts2)
	assert.NoError(t, err)

	val1, err := c.Get(context.Background(), ts1.CacheKey())
	assert.NoError(t, err)
	assert.Equal(t, uint(1), val1.ID)

	val2, err := c.Get(context.Background(), ts2.CacheKey())
	assert.NoError(t, err)
	assert.Equal(t, uint(2), val2.ID)

	err = c.Delete(context.Background(), ts1.CacheKey(), ts2.CacheKey())
	assert.NoError(t, err)

	_, err = c.Get(context.Background(), ts1.CacheKey())
	assert.Error(t, err)

	_, err = c.Get(context.Background(), ts2.CacheKey())
	assert.Error(t, err)
}

func TestCacheIncr(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return nil, nil
	})
	assert.NoError(t, err)

	key := "rate_limit_test"
	_ = c.Delete(context.Background(), key)

	count, err := c.Incr(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = c.Incr(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	count, err = c.Incr(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	_ = c.Delete(context.Background(), key)
}

func TestCacheIncrBy(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return nil, nil
	})
	assert.NoError(t, err)

	key := "incr_by_test"
	_ = c.Delete(context.Background(), key)

	count, err := c.IncrBy(context.Background(), key, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)

	count, err = c.IncrBy(context.Background(), key, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), count)

	count, err = c.IncrBy(context.Background(), key, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(18), count)

	_ = c.Delete(context.Background(), key)
}

func TestCacheIncrBy_Decrement(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return nil, nil
	})
	assert.NoError(t, err)

	key := "incr_by_decrement_test"
	_ = c.Delete(context.Background(), key)

	count, err := c.IncrBy(context.Background(), key, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)

	count, err = c.IncrBy(context.Background(), key, -3)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), count)

	count, err = c.IncrBy(context.Background(), key, -10)
	assert.NoError(t, err)
	assert.Equal(t, int64(-3), count)

	_ = c.Delete(context.Background(), key)
}

func TestCacheExpire(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return nil, nil
	})
	assert.NoError(t, err)

	key := "expire_test"
	_ = c.Delete(context.Background(), key)

	_, err = c.Incr(context.Background(), key)
	assert.NoError(t, err)

	ok, err := c.Expire(context.Background(), key, 1*time.Second)
	assert.NoError(t, err)
	assert.True(t, ok)

	count, err := c.Incr(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	time.Sleep(1100 * time.Millisecond)

	count, err = c.Incr(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_ = c.Delete(context.Background(), key)
}

func TestCacheExpire_NonExistentKey(t *testing.T) {
	c, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return nil, nil
	})
	assert.NoError(t, err)

	ok, err := c.Expire(context.Background(), "non_existent_key_12345", 1*time.Minute)
	assert.NoError(t, err)
	assert.False(t, ok)
}
