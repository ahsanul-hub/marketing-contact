package database

import (
	"app/config"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: config.Config("REDIS_PASS", ""),
		DB:       0,
	})

	// Test connection
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️  Redis connection failed: %v", err)
		log.Printf("💡 Make sure Redis server is running: redis-server")
	} else {
		log.Printf("✅ Redis connection successful")
	}
}
