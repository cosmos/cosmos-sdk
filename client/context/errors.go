package context

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
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

// ErrInvalidSigner is returned when an improper key tries to sign the message
func ErrInvalidSigner(signer sdk.AccAddress, signers []sdk.AccAddress) error {
	return errors.Errorf(`"The generated transaction's intended signer(s) [%v] does not match the given signer: %s"`, signers, signer)
}

// ErrInsufficientFunds is returned when a transaction is attempted from an
// account with insufficient funds
func ErrInsufficientFunds(account auth.Account, amount sdk.Coins) error {
	return errors.Errorf("Address %s has %s coins, transaction requires %s", account.GetAddress(), account.GetCoins(), amount)
}
