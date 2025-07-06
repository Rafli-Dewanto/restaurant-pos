package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)


type RedisCache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

type RedisCacheService struct {
	client *redis.Client
}

func NewRedisCacheService(ctx context.Context, redisURI string) *RedisCacheService {
	opt, _ := redis.ParseURL(redisURI)
	rdb := redis.NewClient(opt)

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully!")
	return &RedisCacheService{client: rdb}
}

// Get retrieves data from Redis and unmarshals it into dest.
func (s *RedisCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found in cache: %s", key)
		}
		return fmt.Errorf("failed to get from Redis: %w", err)
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores data in Redis with an expiration.
func (s *RedisCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for Redis: %w", err)
	}
	return s.client.Set(ctx, key, data, expiration).Err()
}

// Delete removes a key from Redis.
func (s *RedisCacheService) Delete(ctx context.Context, key string) error {
	log.Printf("Cache: Invalidating key: %s", key)
	return s.client.Del(ctx, key).Err()
}
