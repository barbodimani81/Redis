package store

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func Ping(ctx context.Context, client *redis.Client) {
	start := time.Now()
	pong, err := client.Ping(ctx).Result()
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("redis PING failed: %v", err)
	}
	log.Printf("PING response: %s (took %v)\n", pong, elapsed)
}

func Flush(ctx context.Context, client *redis.Client) {
	if err := client.FlushDB(ctx).Err(); err != nil {
		log.Fatalf("FlushDB failed: %v", err)
	}
}
