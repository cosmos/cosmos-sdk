// Package iavl implements the IAVL tree commitment store.
package iavl

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/migration"
)

// HistoricalQuerier wraps an IAVL tree to support historical queries.
// It provides functionality to query data from either the old or new tree
// based on the migration height and configuration.
type HistoricalQuerier struct {
	tree      *IavlTree
	oldReader store.Reader
	db        store.KVStoreWithBatch
	config    *Config
}

// NewHistoricalQuerier creates a new HistoricalQuerier instance.
// It requires a tree, an optional old reader for historical queries,
// a database connection, and configuration.
// Returns error if required parameters are nil.
func NewHistoricalQuerier(tree *IavlTree, oldReader store.Reader, db store.KVStoreWithBatch, config *Config) (*HistoricalQuerier, error) {
	if tree == nil {
		return nil, fmt.Errorf("tree cannot be nil")
	}
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	if config == nil {
		config = DefaultConfig()
	}
	return &HistoricalQuerier{
		tree:      tree,
		oldReader: oldReader,
		db:        db,
		config:    config,
	}, nil
}

// Get retrieves a value at the specified version.
// If historical queries are enabled and the version is before migration height,
// it will use the old reader. Otherwise, it uses the current tree.
func (h *HistoricalQuerier) Get(version uint64, key []byte) ([]byte, error) {
	if !h.config.EnableHistoricalQueries {
		return h.tree.Get(version, key)
	}

	migrationHeight, err := migration.GetMigrationHeight(h.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration height: %w", err)
	}

	if version < migrationHeight && h.oldReader != nil {
		return h.oldReader.Get(version, key)
	}

	return h.tree.Get(version, key)
}

// Iterator returns an iterator over a domain of keys at the specified version.
// If historical queries are enabled and the version is before migration height,
// it will use the old reader. Otherwise, it uses the current tree.
func (h *HistoricalQuerier) Iterator(version uint64, start, end []byte, ascending bool) (store.Iterator, error) {
	if !h.config.EnableHistoricalQueries {
		return h.tree.Iterator(version, start, end, ascending)
	}

	migrationHeight, err := migration.GetMigrationHeight(h.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration height: %w", err)
	}

	if version < migrationHeight && h.oldReader != nil {
		return h.oldReader.Iterator(version, start, end, ascending)
	}

	return h.tree.Iterator(version, start, end, ascending)
} 
