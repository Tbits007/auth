package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tbits007/auth/internal/domain/models"
	"github.com/Tbits007/auth/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (u *UserRepo) Save(
	ctx context.Context,
	user models.User,
) (uuid.UUID, error) {
	const op = "postgres.userRepo.Save"

	query := `
	INSERT INTO users (email, hashed_password)
	VALUES ($1, $2)
	RETURNING id
	`

	var id uuid.UUID
    var err error

    querier := GetQuerier(ctx, u.db)

    err = querier.QueryRow(ctx, query,
        user.Email,
        user.HashedPassword,
    ).Scan(&id)


    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return uuid.Nil, fmt.Errorf("%s: email already exists: %w", op, storage.ErrUserExists)
        }
        return uuid.Nil, fmt.Errorf("%s: failed to save user: %w", op, err)
    }

	return id, nil
} 


func (u *UserRepo) GetByID(
	ctx context.Context,
	userID uuid.UUID,
) (*models.User, error) {
	const op = "postgres.userRepo.GetByID"

	query := `
	SELECT email, hashed_password, is_admin
	FROM users
	WHERE id = $1
	`

    var user models.User
    querier := GetQuerier(ctx, u.db)
    err := querier.QueryRow(ctx, query, userID).Scan(
        &user.Email,
        &user.HashedPassword,
        &user.IsAdmin,
    )

    switch {
    case errors.Is(err,  pgx.ErrNoRows):
        return nil, fmt.Errorf("%s: user not found: %w", op, storage.ErrUserNotFound)
    case err != nil:
        return nil, fmt.Errorf("%s: failed to get user by ID: %w", op, err)
    default:
        return &user, nil
    }

}

func (u *UserRepo) IsAdmin(
	ctx context.Context,
	userID uuid.UUID,
) (bool, error) {
	const op = "postgres.userRepo.IsAdmin"

	query := `
	SELECT is_admin
	FROM users
	WHERE id = $1
	`

    var isAdmin bool
    querier := GetQuerier(ctx, u.db)
    err := querier.QueryRow(ctx, query, userID).Scan(
        &isAdmin,
    )

    switch {
    case errors.Is(err, pgx.ErrNoRows):
        return false, fmt.Errorf("%s: user not found: %w", op, storage.ErrUserNotFound)
    case err != nil:
        return false, fmt.Errorf("%s: failed to get user by ID: %w", op, err)
    default:
        return isAdmin, nil
    }

}

