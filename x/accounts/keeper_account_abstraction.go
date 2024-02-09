package accounts

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	aa_interface_v1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
)

var (
	// ErrAuthentication is returned when the authentication fails.
	ErrAuthentication = errors.New("authentication failed")
	// ErrBundlerPayment is returned when the bundler payment fails.
	ErrBundlerPayment = errors.New("bundler payment failed")
	// ErrExecution is returned when the execution fails.
	ErrExecution = errors.New("execution failed")
	// ErrDisallowedTxCompatInBundle is returned when the tx compat
	// is populated in a bundle.
	ErrDisallowedTxCompatInBundle = errors.New("tx compat field populated in bundle")
)

// IsAbstractedAccount returns if the provided address is an abstracted account or not.
func (k Keeper) IsAbstractedAccount(ctx context.Context, addr []byte) (bool, error) {
	accType, err := k.AccountsByType.Get(ctx, addr)
	switch {
	case errors.Is(err, collections.ErrNotFound):
		return false, nil
	case err != nil:
		return false, err
	}

	impl, err := k.getImplementation(accType)
	if err != nil {
		return false, err
	}
	return impl.HasExec(&aa_interface_v1.MsgAuthenticate{}), nil
}

func (k Keeper) AuthenticateAccount(ctx context.Context) error {

}
