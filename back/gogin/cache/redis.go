package cache

import (
	"context"
	"gogin/config"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
)

func NewRedis() *redis.Client {
	addr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		PoolSize:     200,
		MinIdleConns: 20,
		MaxRetries:   5,
	})

	if err := pingRedis(rdb); err != nil {
		log.Printf("Warning: Redis connection failed: %v. Running without cache.", err)
		_ = rdb.Close()
		return nil
	}

	log.Println("Redis connected successfully")
	return rdb
}

func pingRedis(rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(Ctx, 2*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}
