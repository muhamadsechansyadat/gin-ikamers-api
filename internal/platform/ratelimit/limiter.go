package ratelimit

import (
	"context"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"time"
)

type Limiter struct {
	inner *redis_rate.Limiter
}

func New(rdb *redis.Client) *Limiter {
	return &Limiter{inner: redis_rate.NewLimiter(rdb)}
}

type Result struct {
	Allowed    bool
	Remaining  int
	Limit      int
	RetryAfter time.Duration
}

func (l *Limiter) Allow(ctx context.Context, key string, limit redis_rate.Limit) (*Result, error) {
	r, err := l.inner.Allow(ctx, key, limit)
	if err != nil {
		return nil, err
	}

	return &Result{
		Allowed:    r.Allowed > 0,
		Remaining:  r.Remaining,
		Limit:      r.Limit.Rate,
		RetryAfter: r.RetryAfter,
	}, nil
}

func PerMinute(rate int) redis_rate.Limit {
	return redis_rate.PerMinute(rate)
}

func PerHour(rate int) redis_rate.Limit {
	return redis_rate.PerHour(rate)
}

func Per(rate int, period time.Duration) redis_rate.Limit {
	return redis_rate.Limit{
		Rate:   rate,
		Period: period,
		Burst:  rate,
	}
}

func (l *Limiter) Reset(ctx context.Context, key string) error {
	return l.inner.Reset(ctx, key)
}
