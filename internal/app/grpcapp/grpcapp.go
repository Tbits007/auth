package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/Tbits007/auth/internal/handlers/grpc/auth"
	"github.com/Tbits007/auth/internal/lib/logger/sl"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
    log             *slog.Logger
    gRPCServer      *grpc.Server
    metricsServer   *http.Server
    reg             *prometheus.Registry
    port             int 
} 

func NewGRPCApp(
    log           *slog.Logger, 
    rateLimiter    ratelimit.Limiter, 
    authService    auth.AuthService,
    metricsServer *http.Server,
    reg           *prometheus.Registry,
    port           int,
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

    srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),        
    )
	reg.MustRegister(
        srvMetrics,
    )

    gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
        srvMetrics.UnaryServerInterceptor(),
        logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
        recovery.UnaryServerInterceptor(recoveryOpts...),   
        // ratelimit.UnaryServerInterceptor(rateLimiter),  # depends on Redis
    ))

    auth.NewAuthServer(gRPCServer, authService)

    return &GRPCApp{
        log:           log,
        gRPCServer:    gRPCServer,
        metricsServer: metricsServer,
        reg:           reg,
        port:          port,
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

    go func() {
		m := http.NewServeMux()
		m.Handle("/metrics", promhttp.HandlerFor(
			ga.reg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		ga.metricsServer.Handler = m
		ga.log.Info("starting HTTP server for prometheus")
		if err := ga.metricsServer.ListenAndServe(); err != nil {
            ga.log.Error("failed to start HTTP server for prometheus", sl.Err(err))
        }
    }()

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

    var wg *sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        ga.gRPCServer.GracefulStop()
    }()

    go func () {
        defer wg.Done()
        if err := ga.metricsServer.Shutdown(shutdownCtx); err != nil {
			ga.log.Error("failed to stop web server", sl.Err(err))
		}
    }()

    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        ga.log.Info("gRPC server stopped")
    case <-shutdownCtx.Done():
        ga.log.Info("forcing gRPC server stop")
    }
}