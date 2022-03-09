package store

import (
	"github.com/go-redis/redis"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocksStore "github.com/zhiting-tech/smartassistant/pkg/cache/test/mocks/store/clients"
	"testing"
	"time"
)

func TestNewRedis(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	options := &Options{
		Expiration: 10 * time.Second,
	}

	store := NewRedis(client, options)

	assert.IsType(t, new(RedisStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)

}

func TestRedisGet(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().Get("my-key").Return(&redis.StringCmd{})

	store := NewRedis(client, nil)

	value, err := store.Get("my-key")

	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestRedisSetNX(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().SetNX(cacheKey, cacheValue, 10*time.Second).Return(redis.NewBoolResult(true, nil))

	store := NewRedis(client, nil)

	ok := store.SetNX(cacheKey, cacheValue, 10*time.Second)

	assert.Equal(t, true, ok)
}

func TestRedisSetNXWhenExists(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().SetNX(cacheKey, cacheValue, 10*time.Second).Return(redis.NewBoolResult(false, nil))

	store := NewRedis(client, nil)

	ok := store.SetNX(cacheKey, cacheValue, 10*time.Second)

	assert.Equal(t, false, ok)
}

func TestRedisSet(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{
		Expiration: 6 * time.Second,
	}

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set("my-key", cacheValue, 5*time.Second).Return(&redis.StatusCmd{})

	store := NewRedis(client, options)

	err := store.Set(cacheKey, cacheValue, 5*time.Second)

	assert.Nil(t, err)
}

func TestRedisClear(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().FlushAll().Return(&redis.StatusCmd{})

	store := NewRedis(client, nil)

	err := store.Clear()

	assert.Nil(t, err)
}

func TestRedisGetType(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocksStore.NewMockRedisClientInterface(ctrl)

	store := NewRedis(client, nil)

	assert.Equal(t, RedisType, store.GetType())
}
