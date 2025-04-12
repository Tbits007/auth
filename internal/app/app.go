package app

import (
	"log/slog"
	"time"

	"github.com/Tbits007/auth/internal/app/grpcapp"
	"github.com/Tbits007/auth/internal/services/auth"
	"github.com/Tbits007/auth/internal/storage/postgres"
	"github.com/Tbits007/auth/internal/storage/redis_"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	GRPCServer *grpcapp.GRPCApp
}

func NewApp(
	log 	  *slog.Logger,
	db 		  *pgxpool.Pool,
	rdb		  *redis.Client,
	secretKey  string,
	grpcPort   int,
	tokenTTL   time.Duration,
) *App {

	txManager := postgres.NewTxManager(db)
	userRepo := postgres.NewUserRepo(db)
	eventRepo := postgres.NewEventRepo(db)
	cacheRepo := redis_.NewCacheRepo(rdb)

	authService := auth.NewAuthService(
		log,
		txManager,
		userRepo,
		eventRepo,
		cacheRepo,
		tokenTTL,
		secretKey,
	)

	grpcApp := grpcapp.NewGRPCApp(
		log,
		authService,
		grpcPort,
	)

	return &App{
		GRPCServer: grpcApp,
	}
}