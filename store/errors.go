package store

import (
	"fmt"

	"cosmossdk.io/errors"
)

// StoreCodespace defines the store package's unique error code space.
const StoreCodespace = "store"

var (
	// ErrInvalidProof is returned when a proof is invalid
	ErrInvalidProof = errors.Register(StoreCodespace, 2, "invalid proof")

	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = errors.Register(StoreCodespace, 3, "tx parse error")

	// ErrUnknownRequest to doc
	ErrUnknownRequest = errors.Register(StoreCodespace, 4, "unknown request")

	// ErrLogic defines an internal logic error, e.g. an invariant or assertion
	// that is violated. It is a programmer error, not a user-facing error.
	ErrLogic = errors.Register(StoreCodespace, 5, "internal logic error")

	// ErrConflict defines a conflict error, e.g. when two goroutines try to access
	// the same resource and one of them fails.
	ErrConflict = errors.Register(StoreCodespace, 6, "conflict")

	// ErrInvalidRequest defines an ABCI typed error where the request contains
	// invalid data.
	ErrInvalidRequest = errors.Register(StoreCodespace, 7, "invalid request")

	ErrClosed          = errors.Register(StoreCodespace, 8, "closed")
	ErrRecordNotFound  = errors.Register(StoreCodespace, 9, "record not found")
	ErrUnknownStoreKey = errors.Register(StoreCodespace, 10, "unknown store key")
	ErrKeyEmpty        = errors.Register(StoreCodespace, 11, "key empty")
	ErrStartAfterEnd   = errors.Register(StoreCodespace, 12, "start key after end key")
)

// ErrVersionPruned defines an error returned when a version queried is pruned
// or does not exist.
type ErrVersionPruned struct {
	EarliestVersion uint64
}

func (e ErrVersionPruned) Error() string {
	return fmt.Sprintf("requested version is pruned; earliest available version is: %d", e.EarliestVersion)
}
