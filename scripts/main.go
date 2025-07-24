package main

import (
	"context"
	"fmt"
	"time"

	"github.com/egreerdp/cache"
)

type User struct {
	ID   uint
	Name string
}

func (u *User) Key() string {
	return fmt.Sprintf("%d_%s", u.ID, u.Name)
}

func main() {
	c := cache.NewCache("localhost:6379", "user_cache", 1*time.Minute, func(key string) (*User, error) {
		return &User{
			ID:   2,
			Name: "CallBackUser",
		}, nil
	})

	user := &User{
		ID:   1,
		Name: "User1",
	}

	c.Set(context.TODO(), user)

	u, err := c.Get(context.TODO(), user.Key())
	if err != nil {
		panic(err)
	}

	fmt.Println(u)
}
