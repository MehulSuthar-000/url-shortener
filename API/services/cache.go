package services

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

func CreateClient(dbNo int) *redis.Client {
	redis_addr := os.Getenv("REDIS_ADDR")
	if redis_addr == "" {
		redis_addr = "caching:6379"
	}
	redis_pass := os.Getenv("REDIS_PASS")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_addr,
		Password: redis_pass,
		DB:       dbNo,
	})

	// Ping to verify connection
	_, err := rdb.Ping(Ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis at %s: %v", redis_addr, err)
		// Consider how you want to handle connection failures
		// You might want to panic, return nil, or use a fallback
	}

	return rdb
}
