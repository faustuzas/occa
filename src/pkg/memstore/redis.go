package memstore

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	pkgslices "github.com/faustuzas/occa/src/pkg/slices"
)

type RedisClient struct {
	c *redis.Client
}

func (c RedisClient) SetCollectionItemWithTTL(ctx context.Context, collection string, key string, value []byte, ttl time.Duration) error {
	return c.c.Set(ctx, collection+":"+key, value, ttl).Err()
}

func (c RedisClient) ListCollectionKeys(ctx context.Context, collection string) ([][]byte, error) {
	strResult, err := c.c.Keys(ctx, collection+":*").Result()
	if err != nil {
		return nil, fmt.Errorf("listing keys: %w", err)
	}

	return pkgslices.Map(strResult, func(s string) []byte {
		return []byte(s)
	}), nil
}

func (c RedisClient) Close() error {
	return c.c.Close()
}
