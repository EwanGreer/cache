package cache_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/egreerdp/cache"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	ID   uint
	Name string
}

func (t *TestStruct) Key() string { return fmt.Sprintf("%d_%s", t.ID, t.Name) }

func TestCacheInit(t *testing.T) {
	c := cache.NewCache("", "prefix", func(key string) (*TestStruct, error) { return nil, nil })
	assert.NotNil(t, c)
}

// TODO: make this test functional
func TestCacheGet(t *testing.T) {
	c := cache.NewCache("", "prefix", func(key string) (*TestStruct, error) { return nil, nil })
	_, err := c.Get(context.Background(), "test")
	assert.NoError(t, err)
	// assert.NotNil(t, val)
}

func TestCacheStore(t *testing.T) {
	c := cache.NewCache("", "prefix", func(key string) (*TestStruct, error) { return nil, nil })
	err := c.Set(context.Background(), &TestStruct{})
	assert.NoError(t, err)
}
