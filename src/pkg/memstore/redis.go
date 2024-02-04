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

func (c RedisClient) GetCollectionItem(ctx context.Context, collection string, key string) ([]byte, error) {
	strResult, err := c.c.Get(ctx, collectionKey(collection, key)).Result()
	if err != nil {
		return nil, fmt.Errorf("getting item: %w", err)
	}

	return []byte(strResult), err
}

func (c RedisClient) SetCollectionItemWithTTL(ctx context.Context, collection string, key string, value []byte, ttl time.Duration) error {
	return c.c.Set(ctx, collectionKey(collection, key), value, ttl).Err()
}

func (c RedisClient) ListCollectionKeys(ctx context.Context, collection string) ([]string, error) {
	strResult, err := c.c.Keys(ctx, collectionKey(collection, "*")).Result()
	if err != nil {
		return nil, fmt.Errorf("listing keys: %w", err)
	}

	prefixLen := len(collection)
	for i, s := range strResult {
		strResult[i] = s[prefixLen:]
	}

	return strResult, err
}

func (c RedisClient) listCollectionKeys(ctx context.Context, collection string) ([]string, error) {
	strResult, err := c.c.Keys(ctx, collectionKey(collection, "*")).Result()
	if err != nil {
		return nil, fmt.Errorf("listing keys: %w", err)
	}

	return strResult, err
}

func (c RedisClient) ListCollection(ctx context.Context, collection string) ([][]byte, error) {
	keys, err := c.listCollectionKeys(ctx, collection)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	// TODO: probably not good to fetch everything in one go, should do in batches
	values, err := c.c.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("fetching multiple entries: %w", err)
	}

	return pkgslices.FilterMap(values, func(v interface{}) ([]byte, bool) {
		if v == nil {
			return nil, false
		}
		return []byte(v.(string)), true
	}), nil
}

func (c RedisClient) Close() error {
	return c.c.Close()
}

func collectionKey(collection, key string) string {
	return collection + ":" + key
}
