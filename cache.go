package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cacheable represents an item that can be stored or retrieved
type Cacheable interface {
	Key() string
	Prefix() string
}

type RedisCache[T Cacheable] struct {
	cache    *redis.Client
	ttl      time.Duration
	prefix   string
	callBack CallBackFn[T]
}

// CallBackFn defines the function signature for a cache-miss callback.
// It takes a key (string) and returns the data (T) and an error.
type CallBackFn[T Cacheable] func(ctx context.Context, key string) (T, error)

// NewCache returns an instance of Cache[T]
func NewCache[T Cacheable](cacheURL string, ttl time.Duration, callBackFn CallBackFn[T]) RedisCache[T] {
	client := redis.NewClient(&redis.Options{Addr: cacheURL, DB: 0})

	return RedisCache[T]{
		cache:    client,
		prefix:   newInstanceOfT[T]().Prefix(),
		ttl:      ttl,
		callBack: callBackFn,
	}
}

// Get returns a value from the cache. On a miss the callback is executed, the result is stored in the cache and returned
func (c RedisCache[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T

	result := c.cache.Get(ctx, c.formatKey(key))
	if result.Err() != nil {
		res, err := c.callBack(ctx, c.formatKey(key))
		if err != nil {
			return zero, err
		}

		b, err := json.Marshal(res)
		if err != nil {
			return zero, fmt.Errorf("failed to marshal callback result: %w", err)
		}

		err = c.cache.Set(ctx, c.formatKey(key), b, c.ttl).Err()
		if err != nil {
			return zero, fmt.Errorf("failed to store in cache: %w", err)
		}

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

	err = c.cache.Set(ctx, c.formatKey(item.Key()), b, c.ttl).Err()
	if err != nil {
		return nil
	}

	return nil
}

func (c RedisCache[T]) formatKey(key string) string {
	return fmt.Sprintf("%s_%s", c.prefix, key)
}

func newInstanceOfT[T any]() T {
	var t T
	tType := reflect.TypeOf(t)
	if tType.Kind() == reflect.Ptr {
		return reflect.New(tType.Elem()).Interface().(T)
	}
	return reflect.New(tType).Elem().Interface().(T)
}
