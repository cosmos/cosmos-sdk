package gov

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	StartingProposalID int64                 `json:"starting_proposal_id"`
	Deposits           []DepositWithMetadata `json:"deposits"`
	Votes              []VoteWithMetadata    `json:"votes"`
	Proposals          []Proposal            `json:"proposals"`
	DepositProcedure   DepositProcedure      `json:"deposit_period"`
	VotingProcedure    VotingProcedure       `json:"voting_period"`
	TallyingProcedure  TallyingProcedure     `json:"tallying_procedure"`
}

// DepositWithMetadata (just for genesis)
type DepositWithMetadata struct {
	ProposalID int64   `json:"proposal_id"`
	Deposit    Deposit `json:"deposit"`
}

// VoteWithMetadata (just for genesis)
type VoteWithMetadata struct {
	ProposalID int64 `json:"proposal_id"`
	Vote       Vote  `json:"vote"`
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
			MinDeposit:       sdk.Coins{sdk.NewInt64Coin("steak", 10)},
			MaxDepositPeriod: time.Duration(172800) * time.Second,
		},
		VotingProcedure: VotingProcedure{
			VotingPeriod: time.Duration(172800) * time.Second,
		},
		TallyingProcedure: TallyingProcedure{
			Threshold:         sdk.NewDecWithPrec(5, 1),
			Veto:              sdk.NewDecWithPrec(334, 3),
			GovernancePenalty: sdk.NewDecWithPrec(1, 2),
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
	for _, deposit := range data.Deposits {
		k.setDeposit(ctx, deposit.ProposalID, deposit.Deposit.Depositer, deposit.Deposit)
	}
	for _, vote := range data.Votes {
		k.setVote(ctx, vote.ProposalID, vote.Vote.Voter, vote.Vote)
	}
	for _, proposal := range data.Proposals {
		k.SetProposal(ctx, proposal)
	}
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	startingProposalID, _ := k.getNewProposalID(ctx)
	depositProcedure := k.GetDepositProcedure(ctx)
	votingProcedure := k.GetVotingProcedure(ctx)
	tallyingProcedure := k.GetTallyingProcedure(ctx)
	var deposits []DepositWithMetadata
	var votes []VoteWithMetadata
	proposals := k.GetProposalsFiltered(ctx, nil, nil, StatusNil, 0)
	for _, proposal := range proposals {
		proposalID := proposal.GetProposalID()
		depositsIterator := k.GetDeposits(ctx, proposalID)
		for ; depositsIterator.Valid(); depositsIterator.Next() {
			var deposit Deposit
			k.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
			deposits = append(deposits, DepositWithMetadata{proposalID, deposit})
		}
		votesIterator := k.GetVotes(ctx, proposalID)
		for ; votesIterator.Valid(); votesIterator.Next() {
			var vote Vote
			k.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
			votes = append(votes, VoteWithMetadata{proposalID, vote})
		}
	}

	return GenesisState{
		StartingProposalID: startingProposalID,
		Deposits:           deposits,
		Votes:              votes,
		Proposals:          proposals,
		DepositProcedure:   depositProcedure,
		VotingProcedure:    votingProcedure,
		TallyingProcedure:  tallyingProcedure,
	}
}
