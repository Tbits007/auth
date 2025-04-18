package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tbits007/auth/internal/domain/models/eventModel"
	"github.com/Tbits007/auth/internal/domain/models/userModel"
	"github.com/Tbits007/auth/internal/services/auth"
	"github.com/Tbits007/auth/internal/services/auth/tests/mocks"
	"github.com/Tbits007/auth/internal/services/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
    ctx := context.Background()
    testEmail := "test@example.com"
    testPassword := "password123"
	testUserID := uuid.New()

    mockTxManager := mocks.NewMockTxManager(t)
    mockUserRepo := mocks.NewMockUserRepo(t)
    mockEventRepo := mocks.NewMockEventRepo(t)
    mockCacheRepo := mocks.NewMockCacheRepo(t)

    mockTxManager.EXPECT().
        WithTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(ctx context.Context, fn func(context.Context) error) {
			err := fn(ctx)
			assert.NoError(t, err)
		}).
        Return(nil)

    mockUserRepo.EXPECT().
        Save(ctx, mock.MatchedBy(func(user userModel.User) bool {
            err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(testPassword))
            return user.Email == testEmail && err == nil
        })).
        Return(testUserID, nil)

    mockEventRepo.EXPECT().
		Save(ctx, mock.MatchedBy(func(event eventModel.Event) bool {
			return event.Status == eventModel.PENDING
		})).
        Return(uuid.New(), nil)


    service := auth.NewAuthService(
        testutils.Log,
        mockTxManager,
        mockUserRepo,
        mockEventRepo,
        mockCacheRepo,
        time.Hour,
        "secret",
    )

    userID, err := service.Register(ctx, testEmail, testPassword)

    require.NoError(t, err)
    assert.Equal(t, testUserID, userID)
}

func TestRegister_TransactionError(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	expectedErr := errors.New("transaction error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockTxManager.EXPECT().
		WithTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(expectedErr)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	userID, err := service.Register(ctx, testEmail, testPassword)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, userID)
	assert.Contains(t, err.Error(), expectedErr.Error())

	mockUserRepo.AssertNotCalled(t, "Save")
	mockEventRepo.AssertNotCalled(t, "Save")
}


func TestRegister_UserSaveError(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testPassword := "password123"
	expectedErr := errors.New("user save error")

	mockTxManager := mocks.NewMockTxManager(t)
	mockUserRepo := mocks.NewMockUserRepo(t)
	mockEventRepo := mocks.NewMockEventRepo(t)
	mockCacheRepo := mocks.NewMockCacheRepo(t)

	mockTxManager.EXPECT().
		WithTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(ctx context.Context, fn func(context.Context) error) {
			err := fn(ctx)
			assert.Error(t, err)
		}).
		Return(expectedErr)

	mockUserRepo.EXPECT().
		Save(ctx, mock.AnythingOfType("userModel.User")).
		Return(uuid.Nil, expectedErr)

	service := auth.NewAuthService(
		testutils.Log,
		mockTxManager,
		mockUserRepo,
		mockEventRepo,
		mockCacheRepo,
		time.Hour,
		"secret",
	)

	userID, err := service.Register(ctx, testEmail, testPassword)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, userID)
	assert.Contains(t, err.Error(), expectedErr.Error())

	mockEventRepo.AssertNotCalled(t, "Save")
}