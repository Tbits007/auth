package testutils

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Tbits007/auth/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testDB     *pgxpool.Pool
	dbInitOnce  sync.Once
)

func GetTestDB() *pgxpool.Pool {
	dbInitOnce.Do(func() {
		cfg := config.MustLoad()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	
		connString := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.Postgres.User,
			cfg.Postgres.Password,
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.DBName,
		)

		var err error
		testDB, err = pgxpool.New(ctx, connString)
		if err != nil {
			log.Fatalf("init new pool: %v", err)
		}	
	})

	return testDB
}