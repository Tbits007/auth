package eventRepo

import (
	"context"
	"os"
	"testing"

	"github.com/Tbits007/auth/internal/domain/models/eventModel"
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
	repo := NewEventRepo(testDB)
	cleanTable(t)

	event := eventModel.Event{
		EventType: "user_created",
		Payload:   []byte(`{"user_email":"test@test"}`),
		Status:    "pending",
	}

	id, err := repo.Save(ctx, event)

	require.NoError(t, err)
	assertValidUUID(t, id)
	assertEventExists(t, id)
}

func TestSave_InvalidData(t *testing.T) {
    if testing.Short() {
        t.Skip()
    }

    ctx := context.Background()
    repo := NewEventRepo(testDB)
    cleanTable(t)

    _, err := repo.Save(ctx, eventModel.Event{
        EventType: "invalid",
        Payload:   nil,
        Status:    "invalid",
    })

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to save event")
}

func cleanTable(t *testing.T) {
	_, err := testDB.Exec(context.Background(), "TRUNCATE TABLE outbox CASCADE")
	require.NoError(t, err)
}

func assertValidUUID(t *testing.T, id uuid.UUID) {
	assert.NotEqual(t, uuid.Nil, id)
	assert.Regexp(t, `^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`, id.String())
}

func assertEventExists(t *testing.T, id uuid.UUID) {
	var count int
	err := testDB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM outbox WHERE id = $1",
		id,
	).Scan(&count)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
