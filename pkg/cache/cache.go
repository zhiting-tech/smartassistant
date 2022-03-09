package cache

import (
	"crypto"
	"fmt"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
	"reflect"
	"time"
)

// Cache
type Cache struct {
	store store.StoreInterface
}

// New 创建一个Cache对象
func New(store store.StoreInterface) *Cache {
	return &Cache{store: store}
}

func (c *Cache) getCacheKey(key interface{}) string {
	switch key.(type) {
	case string:
		return key.(string)
	default:
		return checksum(key)
	}
}

func checksum(object interface{}) string {
	digester := crypto.MD5.New()
	fmt.Fprint(digester, reflect.TypeOf(object))
	fmt.Fprint(digester, object)
	hash := digester.Sum(nil)

	return fmt.Sprintf("%x", hash)
}

// Get 如果缓存中存在key对应的缓存对象，则返回该对象，否则返回ErrValueNotFound
func (c *Cache) Get(key string) (interface{}, error) {
	cacheKey := c.getCacheKey(key)
	return c.store.Get(cacheKey)
}

// GetWithTTL 如果缓存中存在key对应的缓存对象，则返回该对象和相应的TTL，
// 否则返回ErrValueNotFound
func (c *Cache) GetWithTTL(key string) (interface{}, time.Duration, error) {
	cacheKey := c.getCacheKey(key)
	return c.store.GetWithTTL(cacheKey)
}

// Set 根据key，和过期时间，过期时间为0表示使用初始化缓存存储的过期时间，将value添加缓存中
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	cacheKey := c.getCacheKey(key)
	return c.store.Set(cacheKey, value, expiration)
}

// Delete 根据key将缓存中的对象删除
func (c *Cache) Delete(key string) error {
	cacheKey := c.getCacheKey(key)
	return c.store.Delete(cacheKey)
}

// SetNX 根据key，和过期时间，过期时间为0表示使用初始化缓存存储的过期时间，将value添加缓存中
// 如果缓存中原来不存在key对应的对象则返回true，否则返回flase
func (c *Cache) SetNX(key string, value interface{}, expiration time.Duration) bool {
	cacheKey := c.getCacheKey(key)
	return c.store.SetNX(cacheKey, value, expiration)
}

// Clear 清除缓存的所有数据
func (c *Cache) Clear() error {
	return c.store.Clear()
}

// GetType 返回缓存的存储类型
func (c *Cache) GetType() string {
	return c.store.GetType()
}

// GetStore 返回缓存存储
func (c *Cache) GetStore() store.StoreInterface {
	return c.store
}
