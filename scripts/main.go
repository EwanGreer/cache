package main

import (
	"context"
	"fmt"

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
	c := cache.NewCache("localhost:6379", "user_cache", func(key string) (*User, error) {
		fmt.Println("CalledBack")
		return &User{
			ID:   2,
			Name: "CallBackUser",
		}, nil
	})

	c.Set(context.TODO(), &User{
		ID:   1,
		Name: "User1",
	})

	u, err := c.Get(context.TODO(), "1_User1")
	if err != nil {
		panic(err)
	}

	fmt.Println(u)
}
