package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tbits007/auth/internal/domain/models/userModel"
	"github.com/Tbits007/auth/internal/services/auth"
	"github.com/Tbits007/auth/internal/services/auth/tests/mocks"
	"github.com/Tbits007/auth/internal/services/testutils"
	"github.com/Tbits007/auth/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)


func TestLogin_CacheHit(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	expectedToken := "cached_token"

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testEmail).
		Return(expectedToken, nil)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	token, err := service.Login(ctx, testEmail, testPassword)

	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)

	mockUserRepo.AssertNotCalled(t, "GetByEmail")
	mockEventRepo.AssertNotCalled(t, "Save")
}

func TestLogin_UserNotFound(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	expectedErr := storage.ErrUserNotFound

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testEmail).
		Return("", errors.New("cache miss"))

	mockUserRepo.EXPECT().
		GetByEmail(ctx, testEmail).
		Return(nil, expectedErr)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	token, err := service.Login(ctx, testEmail, testPassword)

	require.Error(t, err)
	assert.Empty(t, token)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)

	mockEventRepo.AssertNotCalled(t, "Save")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "wrong_password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := userModel.User{
		Email:          testEmail,
		HashedPassword: string(hashedPassword),
	}

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testEmail).
		Return("", errors.New("cache miss"))

	mockUserRepo.EXPECT().
		GetByEmail(ctx, testEmail).
		Return(&user, nil)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	token, err := service.Login(ctx, testEmail, testPassword)

	require.Error(t, err)
	assert.Empty(t, token)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)

	mockEventRepo.AssertNotCalled(t, "Save")
}

func TestLogin_UserRepoError(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	expectedErr := errors.New("user repo error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testEmail).
		Return("", errors.New("cache miss"))

	mockUserRepo.EXPECT().
		GetByEmail(ctx, testEmail).
		Return(nil, expectedErr)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	token, err := service.Login(ctx, testEmail, testPassword)

	require.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), expectedErr.Error())

	mockEventRepo.AssertNotCalled(t, "Save")
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	user := userModel.User{
		ID:             uuid.New(),
		Email:          testEmail,
		HashedPassword: string(hashedPassword),
	}

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testEmail).
		Return("", errors.New("cache miss"))

	mockUserRepo.EXPECT().
		GetByEmail(ctx, testEmail).
		Return(&user, nil)

	mockEventRepo.EXPECT().
		Save(ctx, mock.AnythingOfType("eventModel.Event")).
		Return(uuid.New(), nil)

	mockCacheRepo.EXPECT().
		Set(ctx, testEmail, mock.AnythingOfType("string"), time.Hour).
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

	token, err := service.Login(ctx, testEmail, testPassword)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_EventSaveError(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	user := userModel.User{
		ID:             uuid.New(),
		Email:          testEmail,
		HashedPassword: string(hashedPassword),
	}
	expectedErr := errors.New("event save error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testEmail).
		Return("", errors.New("cache miss"))

	mockUserRepo.EXPECT().
		GetByEmail(ctx, testEmail).
		Return(&user, nil)

	mockEventRepo.EXPECT().
		Save(ctx, mock.AnythingOfType("eventModel.Event")).
		Return(uuid.Nil, expectedErr)

	mockCacheRepo.EXPECT().
		Set(ctx, testEmail, mock.AnythingOfType("string"), time.Hour).
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

	token, err := service.Login(ctx, testEmail, testPassword)

	require.NoError(t, err) 
	assert.NotEmpty(t, token)
}

func TestLogin_CacheSetError(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	user := userModel.User{
		ID:             uuid.New(),
		Email:          testEmail,
		HashedPassword: string(hashedPassword),
	}
	cacheErr := errors.New("cache set error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockCacheRepo.EXPECT().
		Get(ctx, testEmail).
		Return("", errors.New("cache miss"))

	mockUserRepo.EXPECT().
		GetByEmail(ctx, testEmail).
		Return(&user, nil)

	mockEventRepo.EXPECT().
		Save(ctx, mock.AnythingOfType("eventModel.Event")).
		Return(uuid.New(), nil)

	mockCacheRepo.EXPECT().
		Set(ctx, testEmail, mock.AnythingOfType("string"), time.Hour).
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

	token, err := service.Login(ctx, testEmail, testPassword)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}