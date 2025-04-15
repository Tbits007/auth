package userRepo

import (
	"context"
	"os"
	"testing"

	"github.com/Tbits007/auth/internal/domain/models/userModel"
	"github.com/Tbits007/auth/internal/storage/postgres/testutils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSave_Success(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	repo := NewUserRepo(testDB)
	cleanTable(t)

	user := userModel.User{
		Email:          "test@example.com",
		HashedPassword: "hashed_password_123",
	}

	id, err := repo.Save(ctx, user)

	require.NoError(t, err)
	assertValidUUID(t, id)
	assertUserExists(t, id, user.Email)
}

func TestSave_DuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	repo := NewUserRepo(testDB)
	cleanTable(t)

	user := userModel.User{
		Email:          "duplicate@test.com",
		HashedPassword: "hashed_password",
	}
	_, err := repo.Save(ctx, user)
	require.NoError(t, err)

	_, err = repo.Save(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already exists")
}

func cleanTable(t *testing.T) {
	_, err := testDB.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)
}

func assertValidUUID(t *testing.T, id uuid.UUID) {
	assert.NotEqual(t, uuid.Nil, id)
	assert.Regexp(t, `^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`, id.String())
}

func assertUserExists(t *testing.T, id uuid.UUID, expectedEmail string) {
	var (
		email          string
		hashedPassword string
	)

	err := testDB.QueryRow(
		context.Background(),
		"SELECT email, hashed_password FROM users WHERE id = $1",
		id,
	).Scan(&email, &hashedPassword)

	require.NoError(t, err)
	assert.Equal(t, expectedEmail, email)
	assert.NotEmpty(t, hashedPassword)
}