package lock

import (
	"context"
	"geektime-go-cache/lock/custom_errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"time"
)

// SingleFlightRedisLock 处理高并发场景
type SingleFlightRedisLock struct {
	client redis.Cmdable
	sf     singleflight.Group
}

func (s *SingleFlightRedisLock) tryGetLock(ctx context.Context, key string, expiration time.Duration) (*RedisLock, error) {
	val := uuid.New().String()

	ok, err := s.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		// 别人抢到锁了
		return nil, custom_errors.ErrLockPreemptFailed(key)
	}

	return &RedisLock{
		key:        key,
		client:     s.client,
		val:        val,
		expiration: expiration,
		closeChan:  make(chan struct{}),
	}, nil
}

func (s *SingleFlightRedisLock) singleFlightLock(ctx context.Context, key string, expiration time.Duration) (*RedisLock, error) {
	for {
		flag := false
		resCh := s.sf.DoChan(key, func() (interface{}, error) {
			flag = true
			return s.tryGetLock(ctx, key, expiration)
		})

		select {
		case res := <-resCh:
			if flag {
				s.sf.Forget(key)
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*RedisLock), nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
