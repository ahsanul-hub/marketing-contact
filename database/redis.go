package database

import (
	"app/config"
	"context"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {
	// Cek apakah Redis memerlukan password
	redisPass := config.Config("REDIS_PASS", "")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: redisPass,
		DB:       0,
	})

	// Test connection
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️  Redis connection failed: %v", err)
		if strings.Contains(err.Error(), "NOAUTH") {
			log.Printf("💡 Redis requires password. Set REDIS_PASS environment variable")
		} else {
			log.Printf("💡 Make sure Redis server is running: redis-server")
		}
	} else {
		log.Printf("✅ Redis connection successful")
	}
}
