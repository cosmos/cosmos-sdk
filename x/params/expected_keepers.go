package params

import sdk "github.com/cosmos/cosmos-sdk/types"

type ProposalKeeper interface {
	SubmitProposal(ctx sdk.Context, content sdk.ProposalContent) sdk.Error
}
