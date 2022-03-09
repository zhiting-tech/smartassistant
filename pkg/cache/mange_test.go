package cache

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
	"testing"
	"time"
)

func TestInitCache(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	redisStore := store.NewRedis(redisClient, nil)

	InitCache(redisStore)

	assert.Equal(t, redisStore, GetStore())
}

func TestGet(t *testing.T) {
	_ = Set("my-key", "my-value", 0)
	value, err := Get("my-key")

	assert.Nil(t, err)
	assert.Equal(t, "my-value", value)
}

func TestGetWithTTL(t *testing.T) {
	_ = Set("my-key", "my-value", 10*time.Second)

	value, ttl, err := GetWithTTL("my-key")

	assert.Nil(t, err)
	assert.Equal(t, "my-value", value)
	assert.GreaterOrEqual(t, 10*time.Second, ttl)
}

func TestDelete(t *testing.T) {
	_ = Set("my-key", "my-value", 0)
	err := Delete("my-key")

	assert.Nil(t, err)
}

func TestGetStore(t *testing.T) {
	s := GetStore()

	assert.IsType(t, new(store.GoCacheStore), s)
}

func TestSetNX(t *testing.T) {
	ok := SetNX("my-key", "my-value", 0)
	assert.Equal(t, true, ok)

	ok = SetNX("my-key", "my-value", 0)
	assert.Equal(t, false, ok)
}
