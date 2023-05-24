package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	prefix = "cache:"
	Nil    = redis.Nil
)

//go:generate moq -out cache_mock.go . Cache
type Cache interface {
	Get(ctx context.Context, key string, tag string) (string, error)
	Set(ctx context.Context, key string, tag string, value interface{}, expiration time.Duration) error
	DelAll(ctx context.Context, tag string) error
}

type cache struct {
	client *redis.Client
}

func New(client *redis.Client) Cache {
	return &cache{
		client: client,
	}
}

func (cache *cache) Get(ctx context.Context, key string, tag string) (string, error) {
	key = formatKey(key, tag)

	result, err := cache.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return result, nil
}

func (cache *cache) Set(ctx context.Context, key string, tag string, value interface{}, expiration time.Duration) error {
	key = formatKey(key, tag)

	err := cache.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (cache *cache) DelAll(ctx context.Context, tag string) error {
	keys := cache.client.Keys(ctx, formatKey("*", tag)).Val()

	pipeline := cache.client.Pipeline()
	for _, key := range keys {
		err := pipeline.Del(ctx, key).Err()
		if err != nil {
			return err
		}
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func formatKey(key string, tag string) string {
	if tag == "" {
		tag = "*"
	}

	return prefix + tag + ":" + key
}
