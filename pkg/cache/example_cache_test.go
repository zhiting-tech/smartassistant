package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
	"log"
)

func ExampleSet() {
	err := Set("my-key", "my-value", 0)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleSetNX() {
	exists := SetNX("my-key", "my-value", 0)
	fmt.Println(exists)
	// Output:
	// true
}

func ExampleGet() {
	err := Set("my-key", "my-value", 0)
	if err != nil {
		log.Fatal(err)
	}
	val, err := Get("my-key")
	if err != nil {
		if err == store.ErrValueNotFound {
			fmt.Println(err.Error())
		} else {
			log.Fatal(err)
		}
	}
	fmt.Println(val)
	// Output:
	// my-value
}

func ExampleGetWithTTL() {
	err := Set("my-key", "my-value", 0)
	if err != nil {
		log.Fatal(err)
	}
	value, ttl, err := GetWithTTL("my-key")
	if err != nil {
		if err == store.ErrValueNotFound {
			fmt.Println(err.Error())
		} else {
			log.Fatal(err)
		}
	}
	fmt.Println(value, ttl)
}

func ExampleDelete() {
	err := Delete("my-key")
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleInitCache() {
	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	redisStore := store.NewRedis(redisClient, nil)
	InitCache(redisStore)
	err := Set("my-key", "my-value", 0)
	if err != nil {
		log.Fatal(err)
	}
}
