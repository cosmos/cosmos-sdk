package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	StartingProposalID int64             `json:"starting_proposalID"`
	DepositProcedure   DepositProcedure  `json:"deposit_period"`
	VotingProcedure    VotingProcedure   `json:"voting_period"`
	TallyingProcedure  TallyingProcedure `json:"tallying_procedure"`
}

func NewGenesisState(startingProposalID int64, dp DepositProcedure, vp VotingProcedure, tp TallyingProcedure) GenesisState {
	return GenesisState{
		StartingProposalID: startingProposalID,
		DepositProcedure:   dp,
		VotingProcedure:    vp,
		TallyingProcedure:  tp,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		StartingProposalID: 1,
		DepositProcedure: DepositProcedure{
			MinDeposit:       sdk.Coins{sdk.NewCoin("steak", 10)},
			MaxDepositPeriod: 200,
		},
		VotingProcedure: VotingProcedure{
			VotingPeriod: 200,
		},
		TallyingProcedure: TallyingProcedure{
			Threshold:         sdk.NewRat(1, 2),
			Veto:              sdk.NewRat(1, 3),
			GovernancePenalty: sdk.NewRat(1, 100),
		},
	}
}

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	err := k.setInitialProposalID(ctx, data.StartingProposalID)
	if err != nil {
		// TODO: Handle this with #870
		panic(err)
	}
	k.setDepositProcedure(ctx, data.DepositProcedure)
	k.setVotingProcedure(ctx, data.VotingProcedure)
	k.setTallyingProcedure(ctx, data.TallyingProcedure)
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	startingProposalID, _ := k.getNewProposalID(ctx)
	depositProcedure := k.GetDepositProcedure(ctx)
	votingProcedure := k.GetVotingProcedure(ctx)
	tallyingProcedure := k.GetTallyingProcedure(ctx)

	return GenesisState{
		StartingProposalID: startingProposalID,
		DepositProcedure:   depositProcedure,
		VotingProcedure:    votingProcedure,
		TallyingProcedure:  tallyingProcedure,
	}
}
