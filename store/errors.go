package store

import (
	"errors"
)

// Sentinel errors.
var (
	ErrClosed          = errors.New("closed")
	ErrRecordNotFound  = errors.New("record not found")
	ErrUnknownStoreKey = errors.New("unknown store key")
	ErrInvalidVersion  = errors.New("invalid version")
	ErrKeyEmpty        = errors.New("key empty")
	ErrStartAfterEnd   = errors.New("start key after end key")
)
