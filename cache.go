package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cacheable represents an item that can be stored or retrieved
type Cacheable interface {
	CacheKey() string
	// CachePrefix must be nil safe - meaning it should return a constant
	CachePrefix() string
}

type RedisCache[T Cacheable] struct {
	client   *redis.Client
	ttl      time.Duration
	prefix   string
	callBack CallBackFn[T]
}

// CallBackFn defines the function signature for a cache-miss callback.
// It takes a key (string) and returns the data (T) and an error.
type CallBackFn[T Cacheable] func(ctx context.Context, key string) (T, error)

// NewCache returns an instance of Cache[T] and an error
func NewCache[T Cacheable](client *redis.Client, ttl time.Duration, callBackFn CallBackFn[T]) (RedisCache[T], error) {
	var t T

	prefix := t.CachePrefix()

	return RedisCache[T]{
		client:   client,
		prefix:   prefix,
		ttl:      ttl,
		callBack: callBackFn,
	}, nil
}

// Get returns a value from the cache. On a miss the callback is executed, the result is stored in the cache and returned
func (c RedisCache[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T

	result := c.client.Get(ctx, c.formatKey(key))
	if result.Err() != nil {
		res, err := c.callBack(ctx, key)
		if err != nil {
			return zero, err
		}

		err = c.set(ctx, res, false)
		if err != nil {
			return zero, err
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
	return c.set(ctx, item, true)
}

func (c RedisCache[T]) Delete(ctx context.Context, key ...string) error {
	var errs error

	for _, k := range key {
		err := c.client.Del(ctx, c.formatKey(k)).Err()
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (c RedisCache[T]) set(ctx context.Context, item T, failOnConnectionError bool) error {
	b, err := json.Marshal(item)
	if err != nil {
		return resolveError(err, failOnConnectionError)
	}

	err = c.client.Set(ctx, c.formatKey(item.CacheKey()), b, c.ttl).Err()
	if err != nil {
		return resolveError(err, failOnConnectionError)
	}

	return nil
}

func resolveError(err error, suppressDNSErrors bool) error {
	if err == nil {
		return nil
	}

	if suppressDNSErrors && strings.Contains(err.Error(), "no such host") {
		return nil
	}

	return err
}

func (c RedisCache[T]) formatKey(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
