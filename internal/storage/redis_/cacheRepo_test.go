package redis_

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Tbits007/auth/internal/storage"
	"github.com/Tbits007/auth/internal/storage/postgres/testutils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testRDB *redis.Client
)

func TestMain(m *testing.M) {
	testRDB = testutils.GetTestRDB()
	defer testRDB.Close()

	code := m.Run()
	os.Exit(code)
}

func TestSetAndGet_Success(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	repo := NewCacheRepo(testRDB)
	ctx := context.Background()

	key := "test_key_" + t.Name()
	value := "test_value"

	err := repo.Set(ctx, key, value, time.Minute)
	require.NoError(t, err)

	result, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, result)
}

func TestGet_KeyNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	repo := NewCacheRepo(testRDB)
	ctx := context.Background()

	key := "non_existent_key_" + t.Name()

	result, err := repo.Get(ctx, key)

	assert.Error(t, err)
	assert.ErrorIs(t, err, storage.ErrKeyNotFound)
	assert.Empty(t, result)
}

func TestSet_Expiration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	repo := NewCacheRepo(testRDB)
	ctx := context.Background()

	key := "expiring_key_" + t.Name()
	value := "temp_value"
	ttl := 1 * time.Second

	err := repo.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	result, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, result)

	time.Sleep(ttl + 100*time.Millisecond)

	_, err = repo.Get(ctx, key)
	assert.ErrorIs(t, err, storage.ErrKeyNotFound)
}

func TestSet_NonStringValue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	
	repo := NewCacheRepo(testRDB)
	ctx := context.Background()

	key := "int_key_" + t.Name()
	value := 42 

	err := repo.Set(ctx, key, value, time.Minute)
	require.NoError(t, err)

	result, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "42", result)
}
