package migration

import (
	"encoding/binary"
	"fmt"

	corestore "cosmossdk.io/core/store"
)

const (
	// migrationHeightKey is the key used to store the migration height
	migrationHeightKey = "m/height"
)

// StoreMigrationHeight stores the height at which the migration occurred
func StoreMigrationHeight(db corestore.KVStoreWithBatch, height uint64) error {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, height)
	return db.Set([]byte(migrationHeightKey), heightBytes)
}

// GetMigrationHeight returns the height at which the migration occurred
// Returns 0 if no migration has occurred
func GetMigrationHeight(db corestore.KVStoreWithBatch) (uint64, error) {
	heightBytes, err := db.Get([]byte(migrationHeightKey))
	if err != nil {
		return 0, fmt.Errorf("failed to get migration height: %w", err)
	}
	if heightBytes == nil {
		return 0, nil
	}
	return binary.BigEndian.Uint64(heightBytes), nil
} 
