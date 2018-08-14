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

////////////////////  iris/cosmos-sdk begin///////////////////////////
func NewIrisGenesisState(minDeposit sdk.Coins, depositPeriod int64, votingPeriod int64) GenesisState {
	return GenesisState{
		StartingProposalID: 1,
		DepositProcedure: DepositProcedure{
			MinDeposit:       minDeposit,
			MaxDepositPeriod: depositPeriod,
		},
		VotingProcedure: VotingProcedure{
			VotingPeriod: votingPeriod,
		},
		TallyingProcedure: TallyingProcedure{
			Threshold:         sdk.NewRat(1, 2),
			Veto:              sdk.NewRat(1, 3),
			GovernancePenalty: sdk.NewRat(1, 100),
		},
	}
}
////////////////////  iris/cosmos-sdk end///////////////////////////

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		StartingProposalID: 1,
		DepositProcedure: DepositProcedure{
			MinDeposit:       sdk.Coins{sdk.Coin{Denom: "steak", Amount: sdk.NewInt(int64(10)).Mul(Pow10(18))}},
			MaxDepositPeriod: 10,
		},
		VotingProcedure: VotingProcedure{
			VotingPeriod: 10,
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

	k.setDepositProcedureDeposit(ctx, data.DepositProcedure.MinDeposit)
	k.setDepositProcedureMaxDepositPeriod(ctx, data.DepositProcedure.MaxDepositPeriod)

	k.setVotingProcedureVotingPeriod(ctx, data.VotingProcedure.VotingPeriod)

	k.setTallyingProcedure(ctx, ParamStoreKeyTallyingProcedureThreshold, data.TallyingProcedure.Threshold)
	k.setTallyingProcedure(ctx, ParamStoreKeyTallyingProcedureVeto, data.TallyingProcedure.Veto)
	k.setTallyingProcedure(ctx, ParamStoreKeyTallyingProcedurePenalty, data.TallyingProcedure.GovernancePenalty)
}

func Pow10(y int) sdk.Int {
	result := sdk.NewInt(1)
	x := sdk.NewInt(10)
	for i := y; i > 0; i >>= 1 {
		if i&1 != 0 {
			result = result.Mul(x)
		}
		x = x.Mul(x)
	}
	return result
}
