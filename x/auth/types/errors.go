package types

import "errors"

var (
	// ErrAccountNotFound defines the error that is thrown when the account is not found
	ErrAccountNotFound = errors.New("account not found")
)
