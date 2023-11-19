package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Configuration struct {
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func (c Configuration) BuildClient() (Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Address,
		Username: c.User,
		Password: c.Password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	return client{c: redisClient}, nil
}

type Client interface {
	PutIntoCollectionWithTTL(ctx context.Context, collection string, key string, value string, ttl time.Duration) error
	ListCollection(ctx context.Context, collection string) (map[string]string, error)
}

type client struct {
	c *redis.Client
}

func (c client) PutIntoCollectionWithTTL(ctx context.Context, collection string, key string, value string, ttl time.Duration) error {
	return c.c.Set(ctx, collection+":"+key, value, ttl).Err()
}

func (c client) ListCollection(ctx context.Context, collection string) (map[string]string, error) {
	keys, err := c.c.Keys(ctx, collection+":*").Result()
	if err != nil {
		return nil, fmt.Errorf("getting keys: %w", err)
	}

	result := make(map[string]string, len(keys))
	for _, k := range keys {
		k = k[len(collection)+1:]
		result[k] = ""
	}

	return result, nil
}
