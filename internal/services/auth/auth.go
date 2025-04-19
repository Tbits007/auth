package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Tbits007/auth/internal/domain/models/eventModel"
	"github.com/Tbits007/auth/internal/domain/models/userModel"
	"github.com/Tbits007/auth/internal/lib/jwt"
	"github.com/Tbits007/auth/internal/lib/logger/sl"
	"github.com/Tbits007/auth/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

type CacheRepo interface {
	Set(
		ctx 	   context.Context,
		key        string,
		value 	   any, 
		expiration time.Duration,
	) error 

	Get(
		ctx context.Context,
		key string,
	) (string, error)		
}

type AuthService struct {
	log       *slog.Logger
	txManager  TxManager
	userRepo   UserRepo
	eventRepo  EventRepo
	cacheRepo  CacheRepo
	tokenTTL   time.Duration
	secretKey  string
}

func NewAuthService(
	log *slog.Logger,
	txManager TxManager,
	userRepo  UserRepo,
	eventRepo EventRepo,
	cacheRepo CacheRepo,
	tokenTTL  time.Duration,
	secretKey  string,
) *AuthService {
	return &AuthService{
		log: 	   log,
		txManager: txManager,
		userRepo:  userRepo,
		eventRepo: eventRepo,
		cacheRepo: cacheRepo,
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
			return uuid.Nil, fmt.Errorf("%s: generate password hash:%w", op, err)
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

	token, err := au.cacheRepo.Get(ctx, email)
	if err != nil {
		log.Debug("cache miss", sl.Err(err))
	} else {
		log.Debug("cache hit")
		return token, nil
	}

    token, err = jwt.NewToken(*user, au.tokenTTL, au.secretKey)
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

	err = au.cacheRepo.Set(ctx, email, token, 1*time.Hour)
	if err != nil {
		log.Debug("failed to cache token", sl.Err(err))
	}

    return token, nil	
}

func (au *AuthService) IsAdmin(
	ctx   context.Context,
	userID uuid.UUID,
) (bool, error) {
    const op = "AuthService.IsAdmin"

    log := au.log.With(
        slog.String("op", op),
    )	

    cacheVal, err := au.cacheRepo.Get(ctx, userID.String())
    if err == nil {
        log.Debug("cache hit")
        switch cacheVal {
        case "true":
            return true, nil
        case "false":
            return false, nil
        default:
            log.Warn("invalid cache value", slog.String("value", cacheVal))
        }
    } else if !errors.Is(err, redis.Nil) {
        log.Debug("cache error", sl.Err(err))
    }

	isAdmin, err := au.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	
    cacheValue := "false"
    if isAdmin {
        cacheValue = "true"
    }

    if err := au.cacheRepo.Set(ctx, userID.String(), cacheValue, 1*time.Hour); err != nil {
        log.Warn("failed to cache admin status", sl.Err(err))
    }

	return isAdmin, nil
}