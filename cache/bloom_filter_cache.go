package cache

import (
	"context"
	"errors"
	"geektime-go-cache/cache/custom_errors"
	"time"
)

// BloomFilter
// 针对问题：缓存穿透——key不存在都达到数据库上
// 解决方案：BloomFilter 判断hasKey, 存在key时访问数据库
// 实现：BloomFilter 包装 readThrough
type BloomFilter struct {
	hasKey func(ctx context.Context, key string) bool
}

type BloomFilterCache struct {
	*ReadThroughCache
	bf BloomFilter
}

func NewBloomFilterCache(cache *ReadThroughCache, bf BloomFilter, expiration time.Duration, loadFunc func(ctx context.Context, key string) (any, error), opts ...ReadThroughCacheOpt) Cache {
	b := &BloomFilterCache{
		ReadThroughCache: NewReadThroughCache(cache, expiration, loadFunc, opts...),
		bf:               bf,
	}

	loadFunc = func(ctx context.Context, key string) (any, error) {
		if b.bf.hasKey(ctx, key) {
			return loadFunc(ctx, key)
		}
		return nil, custom_errors.ErrFieldNotFound(key)
	}

	return b
}

type BloomFilterCacheV1 struct {
	*ReadThroughCache
	bf BloomFilter
}

func (b *BloomFilterCacheV1) Get(ctx context.Context, key string) (any, error) {
	val, err := b.Cache.Get(ctx, key)
	if err != nil && errors.Is(err, custom_errors.ErrFieldNotFound(key)) {
		if b.bf.hasKey(ctx, key) {
			val, err, _ = b.group.Do(key, func() (interface{}, error) {
				val, err = b.LoadFunc(ctx, key)
				if err != nil {
					b.LogFunc(custom_errors.ErrFieldNotFound(key).Error())
				}

				er := b.Set(ctx, key, val, b.expiration)
				if er != nil {
					b.LogFunc(er.Error())
				}
				return val, err
			})
		}
	}
	return val, err
}
