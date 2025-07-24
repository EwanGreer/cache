package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cacheable represents an item that can be stored or retrieved
type Cacheable interface {
	Key() string
}

type RedisCache[T Cacheable] struct {
	cache    *redis.Client
	prefix   string // prefix before the item key
	callBack CallBackFn[T]
}

// CallBackFn defines the function signature for a cache-miss callback.
// It takes a key (string) and returns the data (T) and an error.
type CallBackFn[T Cacheable] func(key string) (T, error)

// NewCache returns an instance of Cache[T]
func NewCache[T Cacheable](cacheURL string, prefix string, callBackFn CallBackFn[T]) RedisCache[T] {
	client := redis.NewClient(&redis.Options{Addr: cacheURL})

	return RedisCache[T]{
		cache:    client,
		prefix:   prefix,
		callBack: callBackFn,
	}
}

// Get returns a value from the cache if there is one
func (c RedisCache[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T
	result := c.cache.Get(ctx, fmt.Sprintf("%s_%s", c.prefix, key))

	if result.Err() != nil {
		res, err := c.callBack(fmt.Sprintf("%s_%s", c.prefix, key))
		if err != nil {
			return zero, err
		}

		_ = c.cache.Set(ctx, key, result, 1*time.Minute)
		return res, nil
	}

	var data T
	err := json.Unmarshal([]byte(result.Val()), &data)
	if err != nil {
		return zero, err
	}

	return data, nil
}

// Set saves an item to the cache
func (c RedisCache[T]) Set(ctx context.Context, item T) error {
	b, err := json.Marshal(item)
	if err != nil {
		return err
	}

	_ = c.cache.Set(ctx, fmt.Sprintf("%s_%s", c.prefix, item.Key()), b, 1*time.Minute)

	return nil
}
