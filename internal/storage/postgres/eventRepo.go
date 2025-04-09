package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tbits007/auth/internal/domain/models/eventModel"
	"github.com/Tbits007/auth/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepo struct {
	db *pgxpool.Pool
}

func NewEventRepo(db *pgxpool.Pool) *EventRepo {
	return &EventRepo{
		db: db,
	}
}

func (u *EventRepo) Save(
	ctx context.Context,
	Event eventModel.Event,
) (uuid.UUID, error) {
	const op = "postgres.eventRepo.Save"

	query := `
	INSERT INTO events (event_type, payload, status)
	VALUES ($1, $2, $3)
	RETURNING id
	`

	var id uuid.UUID
    var err error

    querier := GetQuerier(ctx, u.db)

    err = querier.QueryRow(ctx, query,
        Event.EventType,
		Event.Payload,
		Event.Status,
    ).Scan(&id)


    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return uuid.Nil, fmt.Errorf("%s: event already exists: %w", op, storage.ErrEventExists)
        }
        return uuid.Nil, fmt.Errorf("%s: failed to save event: %w", op, err)
    }

	return id, nil
} 

