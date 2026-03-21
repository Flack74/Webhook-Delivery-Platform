package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient() (*redis.Client, error) {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	redisPass := os.Getenv("REDIS_PASS")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redisPass, // no password set
		DB:       0,         // use default DB
	})

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Could not connect to Redis: %v", err)
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Println("Successfully connected to Redis!")
	return rdb, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
