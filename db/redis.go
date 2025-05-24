package db

import (
	"context"
	"fmt"
	"log"

	"ewallet-backend/utils"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	redisHost := utils.GetEnv("REDIS_HOST", "localhost")
	redisPort := utils.GetEnv("REDIS_PORT", "6379")

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	RedisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	err := RedisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Println("✅ Connected to Redis")
}
