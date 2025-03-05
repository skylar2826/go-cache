package cache

import (
	"context"
	"errors"
	"geektime-go-cache/cache/custom_errors"
	"golang.org/x/sync/singleflight"
	"log"
	"time"
)

type ReadThroughCache struct {
	Cache
	LogFunc    func(str string)
	LoadFunc   func(ctx context.Context, key string) (any, error)
	expiration time.Duration
	group      singleflight.Group
}

func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err != nil && errors.Is(err, custom_errors.ErrFieldNotFound(key)) {
		val, err = r.LoadFunc(ctx, key)
		if err != nil {
			return nil, custom_errors.ErrFieldNotFound(key)
		}
		er := r.Set(ctx, key, val, r.expiration)
		if er != nil {
			r.LogFunc(er.Error())
		}
	}

	return val, nil
}

// GetV1 异步
func (r *ReadThroughCache) GetV1(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err != nil && errors.Is(err, custom_errors.ErrFieldNotFound(key)) {
		go func() {
			val, err = r.LoadFunc(ctx, key)
			if err != nil {
				r.LogFunc(custom_errors.ErrFieldNotFound(key).Error())
			}
			er := r.Set(ctx, key, val, r.expiration)
			if er != nil {
				r.LogFunc(er.Error())
			}
		}()
	}
	return val, err
}

// GetV2 半异步
func (r *ReadThroughCache) GetV2(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err != nil && errors.Is(err, custom_errors.ErrFieldNotFound(key)) {
		val, err = r.LoadFunc(ctx, key)
		if err != nil {
			r.LogFunc(custom_errors.ErrFieldNotFound(key).Error())
		}
		go func() {
			er := r.Set(ctx, key, val, r.expiration)
			if er != nil {
				r.LogFunc(er.Error())
			}
		}()
	}
	return val, err
}

// GetV3 + singleFlight
func (r *ReadThroughCache) GetV3(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err != nil && errors.Is(err, custom_errors.ErrFieldNotFound(key)) {
		val, err, _ = r.group.Do(key, func() (interface{}, error) {
			val, err = r.LoadFunc(ctx, key)
			if err != nil {
				r.LogFunc(custom_errors.ErrFieldNotFound(key).Error())
			}

			er := r.Set(ctx, key, val, r.expiration)
			if er != nil {
				r.LogFunc(er.Error())
			}
			return val, err
		})

	}
	return val, err
}

type ReadThroughCacheOpt func(r *ReadThroughCache)

func WithLogFunc(logFunc func(str string)) ReadThroughCacheOpt {
	return func(r *ReadThroughCache) {
		r.LogFunc = logFunc
	}
}

func NewReadThroughCache(cache Cache, expiration time.Duration, loadFunc func(ctx context.Context, key string) (any, error), opts ...ReadThroughCacheOpt) *ReadThroughCache {
	r := &ReadThroughCache{
		Cache:    cache,
		LoadFunc: loadFunc,
		LogFunc: func(str string) {
			log.Printf("ReadThroughCache err: %s\n", str)
		},
		expiration: expiration,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}
