package commitment

import (
	"sync"

	"cosmossdk.io/store/v2/commitment/types"

	ics23 "github.com/cosmos/ics23/go"
)

// Database represents a state commitment store. It is designed to securely store
// and manage the most recent state information, crucial for achieving consensus.
// Each module creates its own instance of Database for managing its specific state.
type Database struct {
	mutex sync.Mutex
	tree  types.Tree
}

// NewDatabase creates a new Database instance.
func NewDatabase(tree types.Tree) *Database {
	return &Database{
		mutex: sync.Mutex{},
		tree:  tree,
	}
}

// WriteBatch writes a batch of key-value pairs to the database.
func (db *Database) WriteBatch(batch *types.Batch) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.tree.WriteBatch(batch)
}

// WorkingHash returns the working hash of the database.
func (db *Database) WorkingHash() []byte {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.tree.WorkingHash()
}

// LoadVersion loads the state at the given version.
func (db *Database) LoadVersion(version uint64) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.tree.LoadVersion(version)
}

// Commit commits the current state to the database.
func (db *Database) Commit() ([]byte, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.tree.Commit()
}

// GetProof returns a proof for the given key and version.
func (db *Database) GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.tree.GetProof(version, key)
}

// GetLatestVersion returns the latest version of the database.
func (db *Database) GetLatestVersion() uint64 {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.tree.GetLatestVersion()
}
