package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, supplyKeeper types.SupplyKeeper, data GenesisState) {

	k.SetProposalID(ctx, data.StartingProposalID)
	k.SetDepositParams(ctx, data.DepositParams)
	k.SetVotingParams(ctx, data.VotingParams)
	k.SetTallyParams(ctx, data.TallyParams)

	// check if the deposits pool account exists
	moduleAcc := k.GetGovernanceAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	var totalDeposits sdk.Coins
	for _, deposit := range data.Deposits {
		k.SetDeposit(ctx, deposit)
		totalDeposits = totalDeposits.Add(deposit.Amount)
	}

	for _, vote := range data.Votes {
		k.SetVote(ctx, vote)
	}

	for _, proposal := range data.Proposals {
		switch proposal.Status {
		case StatusDepositPeriod:
			k.InsertInactiveProposalQueue(ctx, proposal.ProposalID, proposal.DepositEndTime)
		case StatusVotingPeriod:
			k.InsertActiveProposalQueue(ctx, proposal.ProposalID, proposal.VotingEndTime)
		}
		k.SetProposal(ctx, proposal)
	}

	// add coins if not provided on genesis
	if moduleAcc.GetCoins().IsZero() {
		if err := moduleAcc.SetCoins(totalDeposits); err != nil {
			panic(err)
		}
		supplyKeeper.SetModuleAccount(ctx, moduleAcc)
	}
}

// ExportGenesis - output genesis parameters
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	startingProposalID, _ := k.GetProposalID(ctx)
	depositParams := k.GetDepositParams(ctx)
	votingParams := k.GetVotingParams(ctx)
	tallyParams := k.GetTallyParams(ctx)

	proposals := k.GetProposalsFiltered(ctx, nil, nil, StatusNil, 0)

	var proposalsDeposits Deposits
	var proposalsVotes Votes
	for _, proposal := range proposals {
		deposits := k.GetDeposits(ctx, proposal.ProposalID)
		proposalsDeposits = append(proposalsDeposits, deposits...)

		votes := k.GetVotes(ctx, proposal.ProposalID)
		proposalsVotes = append(proposalsVotes, votes...)
	}

	return GenesisState{
		StartingProposalID: startingProposalID,
		Deposits:           proposalsDeposits,
		Votes:              proposalsVotes,
		Proposals:          proposals,
		DepositParams:      depositParams,
		VotingParams:       votingParams,
		TallyParams:        tallyParams,
	}
}

// ValidateGenesis checks if parameters are within valid ranges
func ValidateGenesis(data GenesisState) error {
	threshold := data.TallyParams.Threshold
	if threshold.IsNegative() || threshold.GT(sdk.OneDec()) {
		return fmt.Errorf("Governance vote threshold should be positive and less or equal to one, is %s",
			threshold.String())
	}

	veto := data.TallyParams.Veto
	if veto.IsNegative() || veto.GT(sdk.OneDec()) {
		return fmt.Errorf("Governance vote veto threshold should be positive and less or equal to one, is %s",
			veto.String())
	}

	if !data.DepositParams.MinDeposit.IsValid() {
		return fmt.Errorf("Governance deposit amount must be a valid sdk.Coins amount, is %s",
			data.DepositParams.MinDeposit.String())
	}

	return nil
}
