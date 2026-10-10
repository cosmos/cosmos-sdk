package types

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PubKeyChangeHooks allow modules to reject a pubkey change that would violate
// an invariant involving the account's authentication key.
type PubKeyChangeHooks interface {
	BeforePubKeyChange(ctx context.Context, address sdk.AccAddress, oldPubKey, newPubKey cryptotypes.PubKey) error
}

var _ PubKeyChangeHooks = MultiPubKeyChangeHooks{}

// MultiPubKeyChangeHooks combines pubkey change hooks in registration order.
type MultiPubKeyChangeHooks []PubKeyChangeHooks

// NewMultiPubKeyChangeHooks combines pubkey change hooks in registration order.
func NewMultiPubKeyChangeHooks(hooks ...PubKeyChangeHooks) MultiPubKeyChangeHooks {
	return hooks
}

// BeforePubKeyChange runs each hook and returns the first error.
func (h MultiPubKeyChangeHooks) BeforePubKeyChange(
	ctx context.Context,
	address sdk.AccAddress,
	oldPubKey, newPubKey cryptotypes.PubKey,
) error {
	for i := range h {
		if err := h[i].BeforePubKeyChange(ctx, address, oldPubKey, newPubKey); err != nil {
			return err
		}
	}
	return nil
}
