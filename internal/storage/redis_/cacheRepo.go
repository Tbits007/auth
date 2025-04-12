package redis_

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tbits007/auth/internal/storage"
	"github.com/redis/go-redis/v9"
)

type CacheRepo struct {
	db *redis.Client
}

func NewCacheRepo(db *redis.Client) *CacheRepo {
	return &CacheRepo{db: db}
}

func (ca *CacheRepo) Set(
	ctx 	   context.Context,
	key        string,
	value 	   any, 
	expiration time.Duration,
) error {
	const op = "redis.cacheRepo.Set"

	if err := ca.db.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ca *CacheRepo) Get(
	ctx context.Context,
	key string,
) (string, error) {
	const op = "redis.cacheRepo.Get"

	val, err := ca.db.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("%s: key %s: %w", op, key, storage.ErrKeyNotFound)	
		}		
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return val, nil
}