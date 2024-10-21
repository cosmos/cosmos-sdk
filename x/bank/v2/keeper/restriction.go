package keeper

import (
	"context"

	"cosmossdk.io/x/bank/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sendRestriction is a struct that houses a SendRestrictionFn.
// It exists so that the SendRestrictionFn can be updated in the SendKeeper without needing to have a pointer receiver.
type sendRestriction struct {
	fn types.SendRestrictionFn
}

// newSendRestriction creates a new sendRestriction with nil send restriction.
func newSendRestriction() *sendRestriction {
	return &sendRestriction{
		fn: nil,
	}
}

// append adds the provided restriction to this, to be run after the existing function.
func (r *sendRestriction) append(restriction types.SendRestrictionFn) {
	r.fn = r.fn.Then(restriction)
}

// prepend adds the provided restriction to this, to be run before the existing function.
func (r *sendRestriction) prepend(restriction types.SendRestrictionFn) {
	r.fn = restriction.Then(r.fn)
}

// clear removes the send restriction (sets it to nil).
func (r *sendRestriction) clear() {
	r.fn = nil
}

var _ types.SendRestrictionFn = (*sendRestriction)(nil).apply

// apply applies the send restriction if there is one. If not, it's a no-op.
func (r *sendRestriction) apply(ctx context.Context, fromAddr, toAddr []byte, amt sdk.Coins) ([]byte, error) {
	if r == nil || r.fn == nil {
		return toAddr, nil
	}
	return r.fn(ctx, fromAddr, toAddr, amt)
}

// AppendGlobalSendRestriction adds the provided SendRestrictionFn to run after previously provided global restrictions.
func (k Keeper) AppendGlobalSendRestriction(restriction types.SendRestrictionFn) {
	k.sendRestriction.append(restriction)
}

// PrependGlobalSendRestriction adds the provided SendRestrictionFn to run before previously provided global restrictions.
func (k Keeper) PrependGlobalSendRestriction(restriction types.SendRestrictionFn) {
	k.sendRestriction.prepend(restriction)
}

// ClearGlobalSendRestriction removes the global send restriction (if there is one).
func (k Keeper) ClearGlobalSendRestriction() {
	k.sendRestriction.clear()
}
