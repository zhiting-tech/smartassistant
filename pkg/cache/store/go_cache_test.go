package store

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocksStore "github.com/zhiting-tech/smartassistant/pkg/cache/test/mocks/store/clients"
	"testing"
	"time"
)

func TestNewGoCache(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	options := &Options{Expiration: 10}

	store := NewGoCache(client, options)

	assert.IsType(t, new(GoCacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestGoCacheSetNX(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Add(cacheKey, cacheValue, 10*time.Second).Return(nil)

	store := NewGoCache(client, nil)

	ok := store.SetNX(cacheKey, cacheValue, 10*time.Second)

	assert.Equal(t, true, ok)
}

func TestGoCacheSetNXWhenExists(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Add(cacheKey, cacheValue, 10*time.Second).Return(errors.New("already exists"))

	store := NewGoCache(client, nil)

	ok := store.SetNX(cacheKey, cacheValue, 10*time.Second)

	assert.Equal(t, false, ok)
}

func TestGoCacheGet(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)

	store := NewGoCache(client, nil)

	value, err := store.Get(cacheKey)

	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)

}

func TestGoCacheGetWhenError(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	store := NewGoCache(client, nil)

	value, err := store.Get(cacheKey)

	assert.Nil(t, value)
	assert.Equal(t, ErrValueNotFound, err)
}

func TestGoCacheGetWithTTL(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(cacheValue, time.Now(), true)

	store := NewGoCache(client, nil)

	value, ttl, err := store.GetWithTTL(cacheKey)

	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, int64(0), ttl.Milliseconds())
}

func TestGoCacheGetWithTTLWhenError(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(nil, time.Now(), false)

	store := NewGoCache(client, nil)

	value, ttl, err := store.GetWithTTL(cacheKey)

	assert.Nil(t, value)
	assert.Equal(t, ErrValueNotFound, err)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestGoCacheSet(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{}

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 1*time.Second)

	store := NewGoCache(client, options)

	err := store.Set(cacheKey, cacheValue, 1*time.Second)

	assert.Nil(t, err)
}

func TestGoCacheDelete(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheKey := "my-key"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey)

	store := NewGoCache(client, nil)

	err := store.Delete(cacheKey)

	assert.Nil(t, err)
}

func TestGoCacheClear(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Flush()

	store := NewGoCache(client, nil)

	err := store.Clear()

	assert.Nil(t, err)
}

func TestGoCacheGetType(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)

	store := NewGoCache(client, nil)

	assert.Equal(t, GoCacheType, store.GetType())
}
