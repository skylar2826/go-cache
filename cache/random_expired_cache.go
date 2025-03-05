package cache

import (
	"context"
	"math/rand"
	"time"
)

// RandomExpiredCache
// 针对问题：缓存雪崩——大量key同时过期，打到数据库上
// 解决方案：过期时间添加随机数，使得不同时打到数据库
type RandomExpiredCache struct {
	Cache
}

func (r *RandomExpiredCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	offset := time.Duration(rand.Intn(300)) * time.Second // [0, 300)
	expiration = expiration + offset

	return r.Cache.Set(ctx, key, value, expiration)
}
