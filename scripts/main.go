package main

import (
	"context"
	"fmt"
	"time"

	"github.com/EwanGreer/cache"
	"github.com/redis/go-redis/v9"
)

type User struct {
	ID   uint
	Name string
}

func (u *User) CacheKey() string {
	return fmt.Sprintf("%d_%s", u.ID, u.Name)
}

func (u *User) CachePrefix() string {
	return "user"
}

const (
	cacheTTL = 1 * time.Minute
	cacheURL = "localhost:6379"
)

func main() {
	client := redis.NewClient(&redis.Options{Addr: cacheURL})
	c, err := cache.NewCache(client, cacheTTL, cacheCallback)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	user := &User{
		ID:   1,
		Name: "User1",
	}

	err = c.Set(ctx, user)
	if err != nil {
		panic(err)
	}

	u, err := c.Get(ctx, user.CacheKey())
	if err != nil {
		panic(err)
	}

	fmt.Println(u)
}

func cacheCallback(ctx context.Context, key string) (*User, error) {
	return &User{
		ID:   2,
		Name: "CallbackUser",
	}, nil
}
