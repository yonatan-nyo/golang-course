package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client    *redis.Client
	available bool
}

func NewRedisService(redisAddr, redisPassword string) *RedisService {
	service := &RedisService{
		available: false,
	}

	// Create Redis client with address and password
	service.client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: "default",
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := service.client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Unable to connect to Redis: %v", err)
		return service
	}

	service.available = true
	log.Println("Redis connection established successfully")
	return service
}

// IsAvailable checks if Redis is available
func (rs *RedisService) IsAvailable() bool {
	return rs.available
}

// Set stores a value in Redis with expiration
func (rs *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !rs.available {
		return nil // Skip caching if Redis is not available
	}

	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("Warning: Failed to marshal data for Redis: %v", err)
		return err
	}

	err = rs.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		log.Printf("Warning: Failed to set cache key %s: %v", key, err)
		return err
	}

	return nil
}

// Get retrieves a value from Redis
func (rs *RedisService) Get(ctx context.Context, key string, dest interface{}) error {
	if !rs.available {
		return redis.Nil // Return cache miss if Redis is not available
	}

	data, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("Warning: Failed to get cache key %s: %v", key, err)
		}
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		log.Printf("Warning: Failed to unmarshal cached data for key %s: %v", key, err)
		return err
	}

	return nil
}

// Delete removes a key from Redis
func (rs *RedisService) Delete(ctx context.Context, key string) error {
	if !rs.available {
		return nil // Skip if Redis is not available
	}

	err := rs.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Warning: Failed to delete cache key %s: %v", key, err)
		return err
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (rs *RedisService) DeletePattern(ctx context.Context, pattern string) error {
	if !rs.available {
		return nil // Skip if Redis is not available
	}

	keys, err := rs.client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Warning: Failed to get keys with pattern %s: %v", pattern, err)
		return err
	}

	if len(keys) > 0 {
		err = rs.client.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("Warning: Failed to delete keys with pattern %s: %v", pattern, err)
			return err
		}
	}

	return nil
}

// Close closes the Redis connection
func (rs *RedisService) Close() error {
	if rs.client != nil {
		return rs.client.Close()
	}
	return nil
}
