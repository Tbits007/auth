package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tbits007/auth/internal/services/testutils"
	"github.com/Tbits007/auth/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/Tbits007/auth/internal/services/auth/tests/mocks"
	"github.com/Tbits007/auth/internal/services/auth"
)

func TestIsAdmin_CacheHitTrue(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return("true", nil)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.NoError(t, err)
	assert.True(t, isAdmin)

	mockUserRepo.AssertNotCalled(t, "IsAdmin")
}

func TestIsAdmin_CacheHitFalse(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return("false", nil)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.NoError(t, err)
	assert.False(t, isAdmin)

	mockUserRepo.AssertNotCalled(t, "IsAdmin")
}

func TestIsAdmin_CacheHitInvalid(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	cacheValue := "invalid"

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return(cacheValue, nil)

	mockUserRepo.EXPECT().
		IsAdmin(ctx, testUserID).
		Return(true, nil)

	mockCacheRepo.EXPECT().
		Set(ctx, testUserID.String(), "true", time.Hour).
		Return(nil)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.NoError(t, err)
	assert.True(t, isAdmin)
}

func TestIsAdmin_CacheMiss(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return("", redis.Nil)

	mockUserRepo.EXPECT().
		IsAdmin(ctx, testUserID).
		Return(false, nil)

	mockCacheRepo.EXPECT().
		Set(ctx, testUserID.String(), "false", time.Hour).
		Return(nil)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.NoError(t, err)
	assert.False(t, isAdmin)
}

func TestIsAdmin_UserNotFound(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	expectedErr := storage.ErrUserNotFound

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return("", redis.Nil) 

	mockUserRepo.EXPECT().
		IsAdmin(ctx, testUserID).
		Return(false, expectedErr)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.Error(t, err)
	assert.False(t, isAdmin)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestIsAdmin_UserRepoError(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	expectedErr := errors.New("user repo error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return("", redis.Nil) 

	mockUserRepo.EXPECT().
		IsAdmin(ctx, testUserID).
		Return(false, expectedErr)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.Error(t, err)
	assert.False(t, isAdmin)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestIsAdmin_CacheSetError(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	cacheErr := errors.New("cache set error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return("", redis.Nil) 

	mockUserRepo.EXPECT().
		IsAdmin(ctx, testUserID).
		Return(true, nil)

	mockCacheRepo.EXPECT().
		Set(ctx, testUserID.String(), "true", time.Hour).
		Return(cacheErr)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.NoError(t, err)
	assert.True(t, isAdmin)
}

func TestIsAdmin_CacheErrorNotNil(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	cacheErr := errors.New("cache error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testUserID.String()).
		Return("", cacheErr) 

	mockUserRepo.EXPECT().
		IsAdmin(ctx, testUserID).
		Return(true, nil)

	mockCacheRepo.EXPECT().
		Set(ctx, testUserID.String(), "true", time.Hour).
		Return(nil)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	isAdmin, err := service.IsAdmin(ctx, testUserID)

	require.NoError(t, err)
	assert.True(t, isAdmin)
}