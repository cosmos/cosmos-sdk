package context

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// ErrInvalidAccount returns a standardized error reflecting that a given
// account address does not exist.
func ErrInvalidAccount(addr sdk.AccAddress) error {
	return errors.Errorf(`No account with address %s was found in the state.
Are you sure there has been a transaction involving it?`, addr)
}

// ErrVerifyCommit returns a common error reflecting that the blockchain commit at a given
// height can't be verified. The reason is that the base checkpoint of the certifier is
// newer than the given height
func ErrVerifyCommit(height int64) error {
	return errors.Errorf(`The height of base truststore in gaia-lite is higher than height %d. 
Can't verify blockchain proof at this height. Please set --trust-node to true and try again`, height)
}
