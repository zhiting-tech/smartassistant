package cache

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocksStore "github.com/zhiting-tech/smartassistant/pkg/cache/test/mocks/store"
	"time"

	"testing"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	cache := New(store)

	assert.IsType(t, new(Cache), cache)

	assert.Equal(t, store, cache.GetStore())
}

func TestCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Set("my-key", value, 5*time.Second).Return(nil)

	cache := New(store)

	// When
	err := cache.Set("my-key", value, 5*time.Second)
	assert.Nil(t, err)
}

func TestCacheSetNX(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().SetNX("my-key", value, 5*time.Second).Return(true)

	cache := New(store)

	// When
	ok := cache.Set("my-key", value, 5*time.Second)
	assert.Equal(t, true, ok)
}

func TestCacheSetNXExists(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().SetNX("my-key", value, 5*time.Second).Return(false)

	cache := New(store)

	// When
	ok := cache.SetNX("my-key", value, 5*time.Second)
	assert.Equal(t, false, ok)
}

func TestCacheGet(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get("my-key").Return(cacheValue, nil)

	cache := New(store)

	value, err := cache.Get("my-key")

	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	returnedErr := errors.New("Unable to find item in store")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get("my-key").Return(nil, returnedErr)

	cache := New(store)

	value, err := cache.Get("my-key")

	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
}

func TestCacheGetWithTTL(t *testing.T) {
	ctrl := gomock.NewController(t)

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}
	expiration := 1 * time.Second

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().GetWithTTL("my-key").
		Return(cacheValue, expiration, nil)

	cache := New(store)

	// When
	value, ttl, err := cache.GetWithTTL("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, expiration, ttl)
}

func TestCacheGetWithTTLWhenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	returnedErr := errors.New("Unable to find item in store")
	expiration := 0 * time.Second

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().GetWithTTL("my-key").
		Return(nil, expiration, returnedErr)

	cache := New(store)

	value, ttl, err := cache.GetWithTTL("my-key")

	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
	assert.Equal(t, expiration, ttl)
}

func TestCacheDelete(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Delete("my-key").Return(nil)

	cache := New(store)

	err := cache.Delete("my-key")

	assert.Nil(t, err)
}

func TestCacheClear(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Clear().Return(nil)

	cache := New(store)

	err := cache.Clear()

	assert.Nil(t, err)
}

func TestCacheClearWhenError(t *testing.T) {
	ctrl := gomock.NewController(t)

	expectedErr := errors.New("Unexpected error during invalidation")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Clear().Return(expectedErr)

	cache := New(store)

	err := cache.Clear()

	assert.Equal(t, expectedErr, err)
}
