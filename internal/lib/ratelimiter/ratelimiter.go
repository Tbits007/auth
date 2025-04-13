package ratelimiter

import (
	"context"
	"errors"

	"github.com/go-redis/redis_rate/v10"
)

var (
	ErrRateLimited = errors.New("rate limited")
)

type Limiter struct {
	rateLimit *redis_rate.Limiter
}

func NewLimiter(
	rateLimit *redis_rate.Limiter,
) *Limiter {
	return &Limiter{
		rateLimit: rateLimit,
	}
}

func (li *Limiter) Limit(ctx context.Context) error {
	res, err := li.rateLimit.Allow(ctx, "auth", redis_rate.PerSecond(15))
	if err != nil {
		return err
	}
	if res.Allowed == 0 {
		return ErrRateLimited
	}

	return nil
}
