package auth

import (
	"context"
	"errors"
	"github.com/Tbits007/auth/internal/services/auth"
	"github.com/Tbits007/auth/internal/storage"
	au "github.com/Tbits007/contract/gen/go/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(
		ctx context.Context,
		email string,
		password string,
	) (uuid.UUID, error)
	
	Login(
		ctx context.Context,
		email string,
		password string,  
	) (string, error)

	IsAdmin(
	ctx   context.Context,
	userID uuid.UUID,
	) (bool, error)
}


type AuthServer struct {
	au.UnimplementedAuthServer 
	authService AuthService
}

func NewAuthServer(
	gRPCServer *grpc.Server,
	authService AuthService,
) {
	au.RegisterAuthServer(
		gRPCServer,
		&AuthServer{
			authService: authService,
		},
	)  
}

func (as *AuthServer) Register(
	ctx 	context.Context,
	request *au.RegisterRequest,
) (*au.RegisterResponse, error) {
    if request.Email == "" {
        return nil, status.Error(codes.InvalidArgument, "email is required")
    }

    if request.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "password is required")
    }

	uuid, err := as.authService.Register(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &au.RegisterResponse{UserId: uuid.String()}, nil 
}

func (as *AuthServer) Login(
	ctx     context.Context, 
	request *au.LoginRequest,
) (*au.LoginResponse, error) {
    if request.Email == "" {
        return nil, status.Error(codes.InvalidArgument, "email is required")
    }

    if request.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "password is required")
    }

    token, err := as.authService.Login(ctx, request.GetEmail(), request.GetPassword())
    if err != nil {
        if errors.Is(err, auth.ErrInvalidCredentials) {
            return nil, status.Error(codes.InvalidArgument, "invalid email or password")
        }

        return nil, status.Error(codes.Internal, "failed to login")
    }

    return &au.LoginResponse{Token: token}, nil	
}

func (as *AuthServer) IsAdmin(
	ctx 	context.Context, 
	request *au.IsAdminRequest,
) (*au.IsAdminResponse, error) {
	if request.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

    userID, err := uuid.Parse(request.GetUserId())
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
    }

	isAdmin, err := as.authService.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to check admin status")
	}

	return &au.IsAdminResponse{IsAdmin: isAdmin}, nil	
}