package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidProof is returned when a proof is invalid
	ErrInvalidProof = errors.New("invalid proof")

	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = errors.New("tx parse error")

	// ErrUnknownRequest to doc
	ErrUnknownRequest = errors.New("unknown request")

	// ErrLogic defines an internal logic error, e.g. an invariant or assertion
	// that is violated. It is a programmer error, not a user-facing error.
	ErrLogic = errors.New("internal logic error")

	// ErrConflict defines a conflict error, e.g. when two goroutines try to access
	// the same resource and one of them fails.
	ErrConflict = errors.New("conflict")

	// ErrInvalidRequest defines an ABCI typed error where the request contains
	// invalid data.
	ErrInvalidRequest = errors.New("invalid request")

	ErrClosed          = errors.New("closed")
	ErrRecordNotFound  = errors.New("record not found")
	ErrUnknownStoreKey = errors.New("unknown store key")
	ErrKeyEmpty        = errors.New("key empty")
	ErrStartAfterEnd   = errors.New("start key after end key")

	// ErrBatchClosed is returned when a closed or written batch is used.
	ErrBatchClosed = errors.New("batch has been written or closed")

	// ErrValueNil is returned when attempting to set a nil value.
	ErrValueNil = errors.New("value nil")
)

// ErrVersionPruned defines an error returned when a version queried is pruned
// or does not exist.
type ErrVersionPruned struct {
	RequestedVersion uint64
	EarliestVersion  uint64
}

func (e ErrVersionPruned) Error() string {
	return fmt.Sprintf("requested version %d is pruned; earliest available version is: %d", e.RequestedVersion, e.EarliestVersion)
}
