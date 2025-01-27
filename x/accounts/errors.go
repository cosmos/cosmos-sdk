package accounts

import "cosmossdk.io/errors"

var (
	ErrAASemantics = errors.New(ModuleName, 0, "invalid account abstraction tx semantics")
	// ErrAuthentication is returned when the authentication fails.
	ErrAuthentication = errors.New(ModuleName, 1, "authentication failed")
	// ErrBundlerPayment is returned when the bundler payment fails.
	ErrBundlerPayment = errors.New(ModuleName, 2, "bundler payment failed")
	// ErrExecution is returned when the execution fails.
	ErrExecution = errors.New(ModuleName, 3, "execution failed")
	// ErrAccountAlreadyExists is returned when the account already exists in state.
	ErrAccountAlreadyExists = errors.New(ModuleName, 4, "account already exists")
)
