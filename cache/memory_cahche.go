package cache

import (
	"context"
	"geektime-go-cache/cache/custom_errors"
	"log"
	"sync"
	"time"
)

type item struct {
	val      any
	deadline time.Time
}

func (i *item) isExpired(t time.Time) bool {
	if i.deadline.IsZero() {
		return true
	}
	return i.deadline.Before(t)
}

type BuildInMemoryCache struct {
	mu        sync.RWMutex
	data      map[string]item
	close     chan struct{}
	onEvicted func(key string, val any)
}

func (b *BuildInMemoryCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.set(ctx, key, value, expiration)
}

func (b *BuildInMemoryCache) set(ctx context.Context, key string, value any, expiration time.Duration) error {
	val := item{
		val:      value,
		deadline: time.Now().Add(expiration),
	}

	b.data[key] = val
	return nil
}

func (b *BuildInMemoryCache) Get(ctx context.Context, key string) (any, error) {
	b.mu.RLock()
	val, ok := b.data[key]
	if !ok {
		return nil, custom_errors.ErrFieldNotFound(key)
	}
	b.mu.RUnlock()
	if val.isExpired(time.Now()) {
		err := b.delete(ctx, key)
		if err != nil {
			log.Println(err.Error())
		}

		return nil, custom_errors.ErrFieldNotFound(key)
	}
	return val.val, nil
}

func (b *BuildInMemoryCache) Delete(ctx context.Context, key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delete(ctx, key)
}
func (b *BuildInMemoryCache) delete(ctx context.Context, key string) error {
	val, ok := b.data[key]
	if !ok {
		return custom_errors.ErrFieldNotFound(key)
	}

	delete(b.data, key)
	b.onEvicted(key, val)
	return nil
}

func (b *BuildInMemoryCache) Close() {
	b.close <- struct{}{}
}

func (b *BuildInMemoryCache) checkDataExpire(t time.Time) {
	// 默认仅检查前100个
	b.mu.Lock()
	i := 0
	for key, val := range b.data {
		if i >= 100 {
			break
		}
		isExpired := val.isExpired(t)
		if isExpired {
			err := b.delete(context.Background(), key)
			if err != nil {
				log.Println(err)
			}
		}
		i++
	}
	b.mu.Unlock()
}

type BuildInMemoryCacheOpt func(b *BuildInMemoryCache)

func WithOnEvicted(fn func(key string, val any)) BuildInMemoryCacheOpt {
	return func(b *BuildInMemoryCache) {
		b.onEvicted = fn
	}
}

func newBuildInMemoryCache(opts ...BuildInMemoryCacheOpt) *BuildInMemoryCache {
	b := &BuildInMemoryCache{
		data:  make(map[string]item, 10),
		close: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(b)
	}

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case t := <-ticker.C:
				b.checkDataExpire(t)
			case <-b.close:
				break
			}
		}
	}()

	return b
}

func NewBuildInMemoryCache(opts ...BuildInMemoryCacheOpt) Cache {
	return newBuildInMemoryCache(opts...)
}
