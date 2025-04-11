package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tbits007/auth/internal/handlers/grpc/auth"
	"github.com/Tbits007/auth/internal/lib/logger/sl"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

type GRPCApp struct {
    log        *slog.Logger
    gRPCServer *grpc.Server
    port       int 
} 

func NewGRPCApp(
    log *slog.Logger, 
    authService auth.AuthService, 
    port int,
) *GRPCApp {
    loggingOpts := []logging.Option{
        logging.WithLogOnEvents(
            logging.PayloadReceived, logging.PayloadSent,
        ),
    }    

    recoveryOpts := []recovery.Option{
        recovery.WithRecoveryHandler(func(p any) (err error) {
            log.Error("Recovered from panic", slog.Any("panic", p))
            return status.Errorf(codes.Internal, "internal error")
        }),
    }

    gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
        logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
        recovery.UnaryServerInterceptor(recoveryOpts...),   
    ))

    auth.NewAuthServer(gRPCServer, authService)

    return &GRPCApp{
        log:        log,
        gRPCServer: gRPCServer,
        port:       port,
    }    
}


func (ga *GRPCApp) MustRun() {
    if err := ga.Run(); err != nil {
        ga.log.Error("gRPC server fatal error", sl.Err(err))
        panic(err)
    }
}

func (ga *GRPCApp) Run() error {
    const op = "GRPCApp.Run"

    l, err := net.Listen("tcp", fmt.Sprintf(":%d", ga.port))
    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    ga.log.Info("grpc server starting", slog.String("addr", l.Addr().String()))

    if err := ga.gRPCServer.Serve(l); err != nil {
        return fmt.Errorf("%s: serve: %w", op, err)
    }
    
    return nil
}

func (ga *GRPCApp) Stop(shutdownCtx context.Context) {
    const op = "GRPCApp.Stop"

    ga.log.With(slog.String("op", op)).
        Info("stopping gRPC server", slog.Int("port", ga.port))

    done := make(chan struct{})
    
    go func() {
        defer close(done)
        ga.gRPCServer.GracefulStop()
    }()

    select {
    case <-done:
        ga.log.Info("gRPC server stopped")
    case <-shutdownCtx.Done():
        ga.log.Info("forcing gRPC server stop")
    }
}