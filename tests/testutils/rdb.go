package testutils

import (
	"sync"
	"github.com/Tbits007/auth/internal/config"
	"github.com/redis/go-redis/v9"
)

var (
	testRDB     *redis.Client
	rdbInitOnce  sync.Once
)

func GetTestRDB() *redis.Client {
	rdbInitOnce.Do(func() {
		cfg := config.MustLoad()
		testRDB = redis.NewClient(&redis.Options{
			Addr:         cfg.Redis.Address,
			DB:           cfg.Redis.DB,
			MaxRetries:   cfg.Redis.MaxRetries,
			DialTimeout:  cfg.Redis.DialTimeout,
			ReadTimeout:  cfg.Redis.Timeout,
			WriteTimeout: cfg.Redis.Timeout,
		})	
	})

	return testRDB
}