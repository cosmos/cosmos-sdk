package params

import sdk "github.com/cosmos/cosmos-sdk/types"

type ProposalKeeper interface {
	SubmitProposal(ctx sdk.Context, content sdk.ProposalContent) (id uint64, err sdk.Error)
	AddDeposit(ctx sdk.Context, id uint64, addr sdk.AccAddress, amt sdk.Coins) (err sdk.Error, votingStarted bool)
}
