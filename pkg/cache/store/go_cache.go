package store

import (
	"time"
)

const GoCacheType = "go-cache"

// GoCacheClientInterface represents a github.com/patrickmn/go-cache client
type GoCacheClientInterface interface {
	Get(k string) (interface{}, bool)
	GetWithExpiration(k string) (interface{}, time.Time, bool)
	Set(k string, x interface{}, d time.Duration)
	Delete(k string)
	Add(k string, x interface{}, d time.Duration) error
	Flush()
}

type GoCacheStore struct {
	client  GoCacheClientInterface
	options *Options
}

func NewGoCache(client GoCacheClientInterface, options *Options) *GoCacheStore {
	if options == nil {
		options = &Options{}
	}

	return &GoCacheStore{
		client:  client,
		options: options,
	}
}

func (s *GoCacheStore) Get(key string) (interface{}, error) {
	var err error
	value, exists := s.client.Get(key)
	if !exists {
		err = ErrValueNotFound
	}

	return value, err
}

func (s *GoCacheStore) GetWithTTL(key string) (interface{}, time.Duration, error) {
	data, t, exists := s.client.GetWithExpiration(key)
	if !exists {
		return data, 0, ErrValueNotFound
	}
	duration := t.Sub(time.Now())
	return data, duration, nil
}

func (s *GoCacheStore) Set(key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = s.options.Expiration
	}
	s.client.Set(key, value, expiration)
	return nil
}

func (s *GoCacheStore) Delete(key string) error {
	s.client.Delete(key)
	return nil
}

func (s *GoCacheStore) SetNX(key string, value interface{}, expiration time.Duration) bool {
	if expiration == 0 {
		expiration = s.options.Expiration
	}
	err := s.client.Add(key, value, expiration)
	if err != nil {
		return false
	}
	return true
}

func (s *GoCacheStore) Clear() error {
	s.client.Flush()
	return nil
}

func (g *GoCacheStore) GetType() string {
	return GoCacheType
}
