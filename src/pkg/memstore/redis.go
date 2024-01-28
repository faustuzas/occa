package memstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	c *redis.Client
}

func (c RedisClient) SetCollectionItemWithTTL(ctx context.Context, collection string, key string, value string, ttl time.Duration) error {
	return c.c.Set(ctx, collection+":"+key, value, ttl).Err()
}

func (c RedisClient) ListCollectionKeys(ctx context.Context, collection string) ([]string, error) {
	return c.c.Keys(ctx, collection+":*").Result()
}

func (c RedisClient) Close() error {
	return c.c.Close()
}
