package database

import "github.com/go-redis/redis/v8"

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // ganti sesuai host
		Password: "redis123",       // isi kalau pakai password
		DB:       0,                // pakai DB index 0
	})
}
