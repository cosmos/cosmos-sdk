package db

import "errors"

var (
	// ErrTransactionClosed is returned when a closed or written transaction is used.
	ErrTransactionClosed = errors.New("transaction has been written or closed")

	// ErrKeyEmpty is returned when attempting to use an empty or nil key.
	ErrKeyEmpty = errors.New("key cannot be empty")

	// ErrValueNil is returned when attempting to set a nil value.
	ErrValueNil = errors.New("value cannot be nil")

	// ErrVersionDoesNotExist is returned when a DB version does not exist.
	ErrVersionDoesNotExist = errors.New("version does not exist")

	// ErrOpenTransactions is returned when open transactions exist which must
	// be discarded/committed before an operation can complete.
	ErrOpenTransactions = errors.New("open transactions exist")

	// ErrReadOnly is returned when a write operation is attempted on a read-only transaction.
	ErrReadOnly = errors.New("cannot modify read-only transaction")

	// ErrInvalidVersion is returned when an operation attempts to use an invalid version ID.
	ErrInvalidVersion = errors.New("invalid version")
)

// DBConnection represents a connection to a versioned database.
// Records are accessed via transaction objects, and must be safe for concurrent creation
// and read and write access.
// Past versions are only accessible read-only.
type DBConnection interface {
	// Reader opens a read-only transaction at the current working version.
	Reader() DBReader

	// ReaderAt opens a read-only transaction at a specified version.
	// Returns ErrVersionDoesNotExist for invalid versions.
	ReaderAt(uint64) (DBReader, error)

	// ReadWriter opens a read-write transaction at the current version.
	ReadWriter() DBReadWriter

	// Writer opens a write-only transaction at the current version.
	Writer() DBWriter

	// Versions returns all saved versions as an immutable set which is safe for concurrent access.
	Versions() (VersionSet, error)

	// SaveNextVersion saves the current contents of the database and returns the next version ID,
	// which will be `Versions().Last()+1`.
	// Returns an error if any open DBWriter transactions exist.
	// TODO: rename to something more descriptive?
	SaveNextVersion() (uint64, error)

	// SaveVersion attempts to save database at a specific version ID, which must be greater than or
	// equal to what would be returned by `SaveNextVersion`.
	// Returns an error if any open DBWriter transactions exist.
	SaveVersion(uint64) error

	// DeleteVersion deletes a saved version. Returns ErrVersionDoesNotExist for invalid versions.
	DeleteVersion(uint64) error

	// Revert reverts the DB state to the last saved version; if none exist, this clears the DB.
	// Returns an error if any open DBWriter transactions exist.
	Revert() error

	// Close closes the database connection.
	Close() error
}

// DBReader is a read-only transaction interface. It is safe for concurrent access.
// Callers must call Discard when done with the transaction.
//
// Keys cannot be nil or empty, while values cannot be nil. Keys and values should be considered
// read-only, both when returned and when given, and must be copied before they are modified.
type DBReader interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// CONTRACT: key, value readonly []byte
	Get([]byte) ([]byte, error)

	// Has checks if a key exists.
	// CONTRACT: key, value readonly []byte
	Has(key []byte) (bool, error)

	// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
	// Close when done. End is exclusive, and start must be less than end. A nil start iterates
	// from the first key, and a nil end iterates to the last key (inclusive). Empty keys are not
	// valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	Iterator(start, end []byte) (Iterator, error)

	// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
	// must call Close when done. End is exclusive, and start must be less than end. A nil end
	// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
	// Empty keys are not valid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	// TODO: replace with an extra argument to Iterator()?
	ReverseIterator(start, end []byte) (Iterator, error)

	// Discard discards the transaction, invalidating any future operations on it.
	Discard() error
}

// DBWriter is a write-only transaction interface.
// It is safe for concurrent writes, following an optimistic (OCC) strategy, detecting any write
// conflicts and returning an error on commit, rather than locking the DB.
// Callers must call Commit or Discard when done with the transaction.
//
// This can be used to wrap a write-optimized batch object if provided by the backend implementation.
type DBWriter interface {
	// Set sets the value for the given key, replacing it if it already exists.
	// CONTRACT: key, value readonly []byte
	Set([]byte, []byte) error

	// Delete deletes the key, or does nothing if the key does not exist.
	// CONTRACT: key readonly []byte
	Delete([]byte) error

	// Commit flushes pending writes and discards the transaction.
	Commit() error

	// Discard discards the transaction, invalidating any future operations on it.
	Discard() error
}

// DBReadWriter is a transaction interface that allows both reading and writing.
type DBReadWriter interface {
	DBReader
	DBWriter
}

// Iterator represents an iterator over a domain of keys. Callers must call Close when done.
// No writes can happen to a domain while there exists an iterator over it, some backends may take
// out database locks to ensure this will not happen.
//
// Callers must make sure the iterator is valid before calling any methods on it, otherwise
// these methods will panic. This is in part caused by most backend databases using this convention.
// Note that the iterator is invalid on contruction: Next() must be called to initialize it to its
// starting position.
//
// As with DBReader, keys and values should be considered read-only, and must be copied before they are
// modified.
//
// Typical usage:
//
//	var itr Iterator = ...
//	defer itr.Close()
//
//	for itr.Next() {
//	  k, v := itr.Key(); itr.Value()
//	  ...
//	}
//	if err := itr.Error(); err != nil {
//	  ...
//	}
type Iterator interface {
	// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
	// CONTRACT: start, end readonly []byte
	Domain() (start []byte, end []byte)

	// Next moves the iterator to the next key in the database, as defined by order of iteration;
	// returns whether the iterator is valid.
	// Once this function returns false, the iterator remains invalid forever.
	Next() bool

	// Key returns the key at the current position. Panics if the iterator is invalid.
	// CONTRACT: key readonly []byte
	Key() (key []byte)

	// Value returns the value at the current position. Panics if the iterator is invalid.
	// CONTRACT: value readonly []byte
	Value() (value []byte)

	// Error returns the last error encountered by the iterator, if any.
	Error() error

	// Close closes the iterator, relasing any allocated resources.
	Close() error
}

// VersionSet specifies a set of existing versions
type VersionSet interface {
	// Last returns the most recent saved version, or 0 if none.
	Last() uint64
	// Count returns the number of saved versions.
	Count() int
	// Iterator returns an iterator over all saved versions.
	Iterator() VersionIterator
	// Equal returns true iff this set is identical to another.
	Equal(VersionSet) bool
	// Exists returns true if a saved version exists.
	Exists(uint64) bool
}

type VersionIterator interface {
	// Next advances the iterator to the next element.
	// Returns whether the iterator is valid; once invalid, it remains invalid forever.
	Next() bool
	// Value returns the version ID at the current position.
	Value() uint64
}
