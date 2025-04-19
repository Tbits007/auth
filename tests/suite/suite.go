package suite

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/Tbits007/auth/internal/config"
	au "github.com/Tbits007/contract/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
    T          *testing.T                  
    Cfg        *config.Config   
    AuthClient  au.AuthClient
}

func NewSuite(t *testing.T) (context.Context, *Suite) {
    t.Helper()   
    t.Parallel() 

    cfg := config.MustLoad()
 
    ctx, cancelCtx := context.WithTimeout(context.Background(), 60*time.Second)

    t.Cleanup(func() {
        t.Helper()
        cancelCtx()
    })

    grpcAddress := net.JoinHostPort("localhost", strconv.Itoa(cfg.GRPCServer.Port))

    cc, err := grpc.NewClient(
        grpcAddress,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    ) 
    if err != nil {
        t.Fatalf("grpc server connection failed: %v", err)
    }

    authClient := au.NewAuthClient(cc)

    return ctx, &Suite{
        T:          t,
        Cfg:        cfg,
        AuthClient: authClient,
    }
}