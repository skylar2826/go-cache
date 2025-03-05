package cache

import (
	"context"
	"time"
)

type WriteThroughCache struct {
	Cache
	LogFunc    func(str string)
	StoreFunc  func(ctx context.Context, key string, val any) error
	expiration time.Duration
}

// Set 半异步
func (w *WriteThroughCache) Set(ctx context.Context, key string, value any) error {
	err := w.StoreFunc(ctx, key, value)
	if err != nil {
		return err
	}
	go func() {
		err = w.Cache.Set(ctx, key, value, w.expiration)
		if err != nil {
			w.LogFunc(err.Error())
		}
	}()
	return nil
}
