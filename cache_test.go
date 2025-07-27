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

func (t *TestStruct) Key() string    { return fmt.Sprintf("%d_%s", t.ID, t.Name) }
func (t *TestStruct) Prefix() string { return "prefix" }

const connection = "localhost:6379"

func TestCacheInit(t *testing.T) {
	c, close, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) { return nil, nil })
	defer close()
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCacheGet(t *testing.T) {
	c, close, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) { return nil, nil })
	defer close()
	assert.NoError(t, err)

	ts := &TestStruct{
		ID:   1,
		Name: "James",
	}

	err = c.Set(context.Background(), ts)
	assert.NoError(t, err)

	val, err := c.Get(context.Background(), ts.Key())
	assert.NoError(t, err)
	assert.NotNil(t, val)
}

func TestCacheStore_CallBack(t *testing.T) {
	c, close, err := cache.NewCache(redis.NewClient(&redis.Options{Addr: connection}), 1*time.Minute, func(ctx context.Context, key string) (*TestStruct, error) {
		return &TestStruct{
			ID:   2,
			Name: "Ryan",
		}, nil
	})
	defer close()
	assert.NoError(t, err)

	result, err := c.Get(context.Background(), "nothing_here")
	assert.NoError(t, err)

	assert.Equal(t, uint(2), result.ID)
	assert.Equal(t, "Ryan", result.Name)
}
