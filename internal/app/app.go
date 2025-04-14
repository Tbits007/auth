package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Tbits007/auth/internal/app/grpcapp"
	"github.com/Tbits007/auth/internal/lib/ratelimiter"
	"github.com/Tbits007/auth/internal/services/auth"
	"github.com/Tbits007/auth/internal/storage/postgres"
	"github.com/Tbits007/auth/internal/storage/redis_"
	"github.com/go-redis/redis_rate/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type App struct {
	GRPCServer *grpcapp.GRPCApp
}

func NewApp(
	log 	  		*slog.Logger,
	db 		  		*pgxpool.Pool,
	rdb		  		*redis.Client,
	rateLimit 		*redis_rate.Limiter,
	secretKey  		 string,
	grpcPort   		 int,
	tokenTTL   		 time.Duration,
	metricsServer	*http.Server,
	reg				*prometheus.Registry,
) *App {

	txManager := postgres.NewTxManager(db)
	userRepo := postgres.NewUserRepo(db)
	eventRepo := postgres.NewEventRepo(db)
	cacheRepo := redis_.NewCacheRepo(rdb)
	rateLimiter := ratelimiter.NewLimiter(rateLimit)
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
		rateLimiter,
		authService,
        metricsServer,
        reg,
		grpcPort,
	)

	return &App{
		GRPCServer: grpcApp,
	}
}