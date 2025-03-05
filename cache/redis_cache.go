package cache

import (
	"context"
	"geektime-go-cache/cache/custom_errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisCache struct {
	client redis.Cmdable
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	str, err := r.client.Set(ctx, key, value, expiration).Result()
	if err != nil {
		return err
	}
	if str != "OK" {
		return custom_errors.ErrFieldSetFailed(key)
	}
	return err
}

func (r *RedisCache) Get(ctx context.Context, key string) (any, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}

func NewRedisCache(client redis.Cmdable) Cache {
	return &RedisCache{
		client: client,
	}
}
