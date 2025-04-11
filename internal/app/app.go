package app

import (
	"log/slog"
	"time"

	"github.com/Tbits007/auth/internal/app/grpcapp"
	"github.com/Tbits007/auth/internal/services/auth"
	"github.com/Tbits007/auth/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	GRPCServer *grpcapp.GRPCApp
}

func NewApp(
	log 	  *slog.Logger,
	db 		  *pgxpool.Pool,
	secretKey  string,
	grpcPort   int,
	tokenTTL   time.Duration,
) *App {

	txManager := postgres.NewTxManager(db)
	userRepo := postgres.NewUserRepo(db)
	eventRepo := postgres.NewEventRepo(db)

	authService := auth.NewAuthService(
		log,
		txManager,
		userRepo,
		eventRepo,
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