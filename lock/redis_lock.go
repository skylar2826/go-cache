package lock

import (
	"context"
	_ "embed"
	"errors"
	"geektime-go-cache/lock/custom_errors"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	//go:embed lua/unlock.lua
	unlockLua string
	//go:embed lua/refresh.lua
	refreshLua string
)

type RedisLock struct {
	client     redis.Cmdable
	expiration time.Duration
	key        string
	val        any
	closeChan  chan struct{}
}

func (r *RedisLock) Unlock(ctx context.Context) error {
	res, err := r.client.Eval(ctx, unlockLua, []string{r.key}, r.val).Result()
	r.closeChan <- struct{}{}
	if err != nil {
		return err
	}
	if res == 0 {
		return custom_errors.ErrLockIsNotMine(r.key)
	}
	return nil
}

func (r *RedisLock) Refresh(ctx context.Context) error {
	res, err := r.client.Eval(ctx, refreshLua, []string{r.key}, []any{r.val, r.expiration.Seconds()}).Result()
	if err != nil {
		return err
	}
	if res == 0 {
		return custom_errors.ErrLockIsNotMine(r.key)
	}
	return nil
}

func (r *RedisLock) AutoRefresh(expiration time.Duration, interval time.Duration) error {
	timeoutChan := make(chan struct{})
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			// 到期续约
			ctx, cancel := context.WithTimeout(context.Background(), expiration)
			err := r.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				close(timeoutChan)
				return nil
			}
		case <-timeoutChan:
			// 超时重试
			ctx, cancel := context.WithTimeout(context.Background(), expiration)
			err := r.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
			}
			if err != nil {
				close(timeoutChan)
				return nil
			}
		case <-r.closeChan:
			// 关闭服务
			close(timeoutChan)
			return nil
		}
	}
}
