package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Tbits007/auth/internal/app"
	"github.com/Tbits007/auth/internal/config"
	"github.com/Tbits007/auth/internal/lib/logger/sl"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
    envLocal = "local"
    envDev   = "dev"
    envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
    log = log.With(slog.String("env", cfg.Env))

    log.Info("initializing server", slog.Int("port", cfg.GRPCServer.Port))
    log.Debug("logger debug mode enabled")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
        cfg.Postgres.Password,
        cfg.Postgres.Host,
        cfg.Postgres.Port,
        cfg.Postgres.DBName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, connString) 
	if err != nil {
		log.Error("failed to initialize db", sl.Err(err))
		os.Exit(1)
	}

	application := app.NewApp(
		log,
		db,
		cfg.Auth.SecretKey,
		cfg.GRPCServer.Port,
		cfg.Auth.TokenTTL,
	)

	application.GRPCServer.MustRun()
	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	gracefulShutdown(log, application, db)
}

func gracefulShutdown(log *slog.Logger, application *app.App, db *pgxpool.Pool) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	done := make(chan struct{})

	log.Info("starting graceful shutdown...")
		
	var wg sync.WaitGroup
	wg.Add(2)
		
	go func() {
		defer wg.Done()
		application.GRPCServer.Stop(shutdownCtx)
	}()
		
	go func() {
		defer wg.Done()
		db.Close()
	}()

    go func() {
        wg.Wait()
		close(done)
    }()
	
	select {
	case <-done:
		log.Info("shutdown completed successfully")
	case <-shutdownCtx.Done():
		log.Error("shutdown timed out, forcing exit")
		os.Exit(1)
	}	
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)		
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)			
	}

	return log 
}