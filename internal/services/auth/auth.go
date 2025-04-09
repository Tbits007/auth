package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Tbits007/auth/internal/domain/models/userModel"
	"github.com/Tbits007/auth/internal/domain/models/eventModel"
	"github.com/Tbits007/auth/internal/lib/jwt"
	"github.com/Tbits007/auth/internal/lib/logger/sl"
	"github.com/Tbits007/auth/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
)


type UserRepo interface {
	Save(
		ctx context.Context,
		user userModel.User,
	) (uuid.UUID, error)

	GetByEmail(
		ctx context.Context,
		email string,
	) (*userModel.User, error)

	IsAdmin(
		ctx context.Context,
		userID uuid.UUID,
	) (bool, error)
}

type EventRepo interface {
	Save(
		ctx context.Context,
		Event eventModel.Event,
	) (uuid.UUID, error)
}

type TxManager interface {
	WithTransaction(
		ctx context.Context,
		fn func(ctx context.Context) error,
	) error
}

type AuthService struct {
	log       *slog.Logger
	txManager  TxManager
	userRepo   UserRepo
	eventRepo  EventRepo
	tokenTTL   time.Duration
	secretKey  string
}

func NewAuthService(
	log *slog.Logger,
	txManager TxManager,
	userRepo  UserRepo,
	eventRepo EventRepo,
	tokenTTL  time.Duration,
	secretKey  string,
) *AuthService {
	return &AuthService{
		log: 	   log,
		txManager: txManager,
		userRepo:  userRepo,
		eventRepo: eventRepo,
		tokenTTL:  tokenTTL,
		secretKey: secretKey,
	}
}


func (au *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
	) (uuid.UUID, error) {
		const op = "AuthService.Register"

		log := au.log.With(
			slog.String("op", op),
		)		

		passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to generate password hash", sl.Err(err))
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
		
		user := userModel.User{
			Email: email,
			HashedPassword: string(passHash),
		}

		eventPayload := map[string]any{
			"email":     email,
			"action":    "registration",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			log.Error("failed to marshal eventPayload", sl.Err(err))
			return uuid.Nil, fmt.Errorf("%s: marshal event payload: %w", op, err)
		}	

		event := eventModel.Event{
			Payload: payloadBytes,
			Status: eventModel.PENDING,
		}

		var userID uuid.UUID

		err = au.txManager.WithTransaction(ctx, func(ctx context.Context) error {
			userID, err = au.userRepo.Save(ctx, user)
			if err != nil {
				return err 
			}
			_, err = au.eventRepo.Save(ctx, event)
			if err != nil {
				return err 
			}
			return nil 
		})
		if err != nil {
			log.Error("transaction failed", sl.Err(err))
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
		
		return userID, nil
}

func (au *AuthService) Login(
    ctx context.Context,
    email string,
    password string,  
) (string, error) {
    const op = "AuthService.Login"

    log := au.log.With(
        slog.String("op", op),
    )

	user, err := au.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

    if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
        log.Info("invalid credentials", sl.Err(err))
        return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
    }	

    token, err := jwt.NewToken(*user, au.tokenTTL, au.secretKey)
    if err != nil {
        log.Error("failed to generate token", sl.Err(err))
        return "", fmt.Errorf("%s: %w", op, err)
    }

	eventPayload := map[string]any{
		"email":     email,
		"action":    "login",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		log.Error("failed to marshal eventPayload", sl.Err(err))
		return "", fmt.Errorf("%s: marshal event payload: %w", op, err)
	}	

	event := eventModel.Event{
		Payload: payloadBytes,
		Status: eventModel.PENDING,
	}

	_, err = au.eventRepo.Save(ctx, event)
	if err != nil {
		log.Error("failed to save event", sl.Err(err))
	}

    return token, nil	
}