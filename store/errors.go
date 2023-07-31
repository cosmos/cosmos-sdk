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
)
