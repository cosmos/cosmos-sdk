package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	StartingProposalID int64 `json:"starting_proposalID"`
}

func NewGenesisState(startingProposalID int64) GenesisState {
	return GenesisState{
		StartingProposalID: startingProposalID,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		StartingProposalID: 1,
	}
}

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.setInitialProposalID(ctx, data.StartingProposalID)
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	initalProposalID, _ := k.getNewProposalID(ctx)

	return GenesisState{
		initalProposalID,
	}
}
