package cache

import (
	"context"
	"geektime-go-cache/cache/custom_errors"
	"time"
)

type MaxMemoryCache struct {
	*BuildInMemoryCache

	// 缓存数量限制
	maxCnt int64
	// 当前数量
	curCnt int64
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.mu.RLock()
	_, ok := m.data[key]
	m.mu.RUnlock()
	if !ok {
		if m.curCnt+1 >= m.maxCnt {
			return custom_errors.ErrFieldOverMaxSize(custom_errors.ErrFieldSetFailed(key).Error())
		}
		m.mu.Lock()
		err := m.set(ctx, key, value, expiration)
		m.mu.Unlock()
		if err != nil {
			return err
		}
		m.curCnt++
	}

	// 更新过期时间
	return m.set(ctx, key, value, expiration)
}

func NewMaxMemoryCache(maxSize int64, maxCnt int64, opts ...BuildInMemoryCacheOpt) Cache {
	b := newBuildInMemoryCache(opts...)

	return &MaxMemoryCache{
		maxCnt:             maxCnt,
		curCnt:             0,
		BuildInMemoryCache: b,
	}
}
