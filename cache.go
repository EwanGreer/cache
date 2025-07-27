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
	// Prefix must be nil safe - meaning it should return a constant
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

// NewCache returns an instance of Cache[T], a cleanup function and a potential error
func NewCache[T Cacheable](client *redis.Client, ttl time.Duration, callBackFn CallBackFn[T]) (RedisCache[T], func() error, error) {
	var zero RedisCache[T]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return zero, nil, fmt.Errorf("failed to connect to Redis at %s: %w", client.Options().Addr, err)
	}

	var t T
	prefix := t.Prefix()

	return RedisCache[T]{
		cache:    client,
		prefix:   prefix,
		ttl:      ttl,
		callBack: callBackFn,
	}, client.Close, nil
}

// Get returns a value from the cache. On a miss the callback is executed, the result is stored in the cache and returned
func (c RedisCache[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T

	result := c.cache.Get(ctx, c.formatKey(key))
	if result.Err() != nil {
		res, err := c.callBack(ctx, key)
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
		return err
	}

	return nil
}

// ParseRedisURL parses a Redis URL and returns redis.Options
// Supported format: redis[s]://[[user][:password]@]host[:port][/db-number]
func ParseRedisURL(redisURL string) (*redis.Options, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis URL cannot be empty")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL %q: %w", redisURL, err)
	}

	return opts, nil
}

func (c RedisCache[T]) formatKey(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
