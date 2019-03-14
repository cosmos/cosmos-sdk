package proposal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Expected keeper to submit and deposit on proposals
type Keeper interface {
	SubmitProposal(ctx sdk.Context, content Content) (id uint64, err sdk.Error)
	AddDeposit(ctx sdk.Context, id uint64, addr sdk.AccAddress, amt sdk.Coins) (err sdk.Error, votingStarted bool)
}
