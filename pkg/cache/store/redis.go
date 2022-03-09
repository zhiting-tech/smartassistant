package store

import (
	"github.com/go-redis/redis"
	"time"
)

type RedisClientInterface interface {
	Get(key string) *redis.StringCmd
	TTL(key string) *redis.DurationCmd
	Set(key string, values interface{}, expiration time.Duration) *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
	FlushAll() *redis.StatusCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
}

const RedisType = "redis"

type RedisStore struct {
	client  RedisClientInterface
	options *Options
}

func NewRedis(client RedisClientInterface, options *Options) *RedisStore {
	if options == nil {
		options = &Options{}
	}

	return &RedisStore{
		client:  client,
		options: options,
	}
}

func (s *RedisStore) Get(key string) (interface{}, error) {
	object, err := s.client.Get(key).Result()
	if err != nil {
		err = ErrValueNotFound
	}
	return object, err
}

func (s *RedisStore) GetWithTTL(key string) (interface{}, time.Duration, error) {
	object, err := s.client.Get(key).Result()
	if err != nil {
		return nil, 0, ErrValueNotFound
	}

	ttl, err := s.client.TTL(key).Result()
	if err != nil {
		return nil, 0, ErrValueNotFound
	}

	return object, ttl, err
}

func (s *RedisStore) Set(key string, val interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = s.options.Expiration
	}
	return s.client.Set(key, val, expiration).Err()
}

func (s *RedisStore) Delete(key string) error {
	_, err := s.client.Del(key).Result()
	return err
}

func (s *RedisStore) SetNX(key string, value interface{}, expiration time.Duration) bool {
	if expiration == 0 {
		expiration = s.options.Expiration
	}
	val, err := s.client.SetNX(key, value, expiration).Result()
	if err != nil {
		return false
	}
	return val
}

func (s *RedisStore) Clear() error {
	return s.client.FlushAll().Err()
}

func (s *RedisStore) GetType() string {
	return RedisType
}
