package migration

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "cosmossdk.io/store/v2/db"
)

func TestMigrationHeight(t *testing.T) {
	db := dbm.NewMemDB()

	// Initially should return 0
	height, err := GetMigrationHeight(db)
	require.NoError(t, err)
	require.Equal(t, uint64(0), height)

	// Store migration height
	testHeight := uint64(1000)
	err = StoreMigrationHeight(db, testHeight)
	require.NoError(t, err)

	// Verify stored height
	height, err = GetMigrationHeight(db)
	require.NoError(t, err)
	require.Equal(t, testHeight, height)
} 
