package store

import (
	"time"
)

type StoreInterface interface {
	Get(key string) (interface{}, error)
	GetWithTTL(key string) (interface{}, time.Duration, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	Clear() error
	GetType() string
	SetNX(key string, value interface{}, expiration time.Duration) bool
}
