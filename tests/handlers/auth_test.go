package handlers

import (
	"context"
	"os"
	"testing"
	"github.com/Tbits007/auth/tests/suite"
	"github.com/Tbits007/auth/tests/testutils"
	au "github.com/Tbits007/contract/gen/go/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
    testDB *pgxpool.Pool
)

func TestMain(m *testing.M) {
    testDB = testutils.GetTestDB()
    defer testDB.Close()

    code := m.Run()
    os.Exit(code)
}

func TestAuthService_Register(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	
	ctx, s := suite.NewSuite(t)

	cleanTables(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:     "successful registration",
			email:    "test1@example.com",
			password: "password123",
		},
		{
			name:        "empty email",
			email:       "",
			password:    "password123",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "empty password",
			email:       "test2@example.com",
			password:    "",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:     "duplicate registration",
			email:    "duplicate@example.com",
			password: "password123",
			expectError: true,
			errorCode:   codes.AlreadyExists,
		},
	}

	_, err := s.AuthClient.Register(ctx, &au.RegisterRequest{
		Email:    "duplicate@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.AuthClient.Register(ctx, &au.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			if tt.expectError {
				require.Error(t, err)
				status, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, status.Code())
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.GetUserId())
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, s := suite.NewSuite(t)

	cleanTables(t)

	registerResp, err := s.AuthClient.Register(ctx, &au.RegisterRequest{
		Email:    "testuser@example.com",
		Password: "correct_password",
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerResp.GetUserId())

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:     "successful login",
			email:    "testuser@example.com",
			password: "correct_password",
		},
		{
			name:        "wrong password",
			email:       "testuser@example.com",
			password:    "wrong_password",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "non-existent user",
			email:       "nonexistent@example.com",
			password:    "any_password",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "empty email",
			email:       "",
			password:    "password123",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "empty password",
			email:       "testuser@example.com",
			password:    "",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.AuthClient.Login(ctx, &au.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			if tt.expectError {
				t.Log(resp)
				require.Error(t, err)
				status, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, status.Code())
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.GetToken(), "token should not be empty")
			
			assert.Greater(t, len(resp.GetToken()), 100, "token seems too short for JWT")
		})
	}
}

func TestAuthService_IsAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, s := suite.NewSuite(t)

	cleanTables(t)

	regularUser, err := s.AuthClient.Register(ctx, &au.RegisterRequest{
		Email:    "regular@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	adminUser, err := s.AuthClient.Register(ctx, &au.RegisterRequest{
		Email:    "admin@example.com",
		Password: "admin123",
	})
	require.NoError(t, err)

	makeAdmin(t, adminUser.GetUserId())

	tests := []struct {
		name        string
		userID      string
		expected    bool
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:     "regular user is not admin",
			userID:   regularUser.GetUserId(),
			expected: false,
		},
		{
			name:     "admin user is admin",
			userID:   adminUser.GetUserId(),
			expected: true,
		},
		{
			name:        "non-existent user",
			userID:      uuid.New().String(),
			expectError: true,
			errorCode:   codes.NotFound,
		},
		{
			name:        "invalid user id format",
			userID:      "invalid-uuid",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.AuthClient.IsAdmin(ctx, &au.IsAdminRequest{
				UserId: tt.userID,
			})

			if tt.expectError {
				t.Log(err)
				require.Error(t, err)
				status, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, status.Code())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, resp.GetIsAdmin())
		})
	}
}

func makeAdmin(t *testing.T, userID string) {
	uuid, err := uuid.Parse(userID)
	require.NoError(t, err)

	_, err = testDB.Exec(context.Background(),
		"UPDATE users SET is_admin = true WHERE id = $1", uuid)
	require.NoError(t, err)
}

func cleanTables(t *testing.T) {
	_, err := testDB.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)
	
	_, err = testDB.Exec(context.Background(), "TRUNCATE TABLE outbox CASCADE")
	require.NoError(t, err)
}
