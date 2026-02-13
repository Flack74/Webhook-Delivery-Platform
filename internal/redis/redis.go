package redis

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func RedisClient() *redis.Client {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	addr := redisHost +":" + redisPort

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Background context for the ping
	var ctx = context.Background()

	// Verify connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Could not connect to Redis: %v", err)
		return nil
	}

	log.Println("Successfully connected to Redis!")
	return rdb
}
