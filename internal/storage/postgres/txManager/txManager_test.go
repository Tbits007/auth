package txManager

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Tbits007/auth/internal/storage/postgres/testutils"
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

func TestWithTransaction_Commit(t *testing.T) {
    ctx := context.Background()
    tm := NewTxManager(testDB)

    err := tm.WithTransaction(ctx, func(ctx context.Context) error {
        tx, ok := GetTx(ctx)
        require.True(t, ok)
        
        _, err := tx.Exec(ctx, "SELECT 1")
        assert.NoError(t, err)
        return nil
    })

    assert.NoError(t, err)
}

func TestWithTransaction_Rollback(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    _, err := testDB.Exec(ctx, `
        CREATE TABLE test_rollback (
            id SERIAL PRIMARY KEY,
            data TEXT
        );
    `)
    t.Cleanup(func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _, _ = testDB.Exec(ctx, "DROP TABLE IF EXISTS test_rollback CASCADE")
    })

    tm := NewTxManager(testDB)
    testErr := errors.New("test error")

    err = tm.WithTransaction(ctx, func(txCtx context.Context) error {
        tx, ok := GetTx(txCtx)
        require.True(t, ok)

        _, err := tx.Exec(txCtx, "INSERT INTO test_rollback(data) VALUES ('test')")
        require.NoError(t, err)

        return testErr
    })
    require.ErrorIs(t, err, testErr)
  
    var count int
    err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM test_rollback").Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 0, count, "changes must be rolled back")
}

