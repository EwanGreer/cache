package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/egreerdp/cache"
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

	// Delete now formats the key internally, so pass the raw cache key
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
