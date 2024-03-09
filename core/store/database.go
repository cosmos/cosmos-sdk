package store

// DatabaseService provides access to the underlying database for CRUD operations of non-consensus data.
// WARNING: using this api will make your module unprovable for fraud and validity proofs
type DatabaseService interface {
	GetDatabase() NonConsensusStore
}

// NonConsensusStore is a simple key-value store that is used to store non-consensus data.
// Note the non-consensus data is not committed to the blockchain and does not allow iteration
type NonConsensusStore interface {
	// Get returns nil iff key doesn't exist. Errors on nil key.
	Get(key []byte) ([]byte, error)

	// Has checks if a key exists. Errors on nil key.
	Has(key []byte) (bool, error)

	// Set sets the key. Errors on nil key or value.
	Set(key, value []byte) error

	// Delete deletes the key. Errors on nil key.
	Delete(key []byte) error
}
