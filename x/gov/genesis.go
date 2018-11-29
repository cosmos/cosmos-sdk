package gov

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	StartingProposalID uint64                `json:"starting_proposal_id"`
	Deposits           []DepositWithMetadata `json:"deposits"`
	Votes              []VoteWithMetadata    `json:"votes"`
	Proposals          []Proposal            `json:"proposals"`
	DepositParams      DepositParams         `json:"deposit_params"`
	VotingParams       VotingParams          `json:"voting_params"`
	TallyParams        TallyParams           `json:"tally_params"`
}

// DepositWithMetadata (just for genesis)
type DepositWithMetadata struct {
	ProposalID uint64  `json:"proposal_id"`
	Deposit    Deposit `json:"deposit"`
}

// VoteWithMetadata (just for genesis)
type VoteWithMetadata struct {
	ProposalID uint64 `json:"proposal_id"`
	Vote       Vote   `json:"vote"`
}

func NewGenesisState(startingProposalID uint64, dp DepositParams, vp VotingParams, tp TallyParams) GenesisState {
	return GenesisState{
		StartingProposalID: startingProposalID,
		DepositParams:      dp,
		VotingParams:       vp,
		TallyParams:        tp,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		StartingProposalID: 1,
		DepositParams: DepositParams{
			MinDeposit:       sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 10)},
			MaxDepositPeriod: time.Duration(172800) * time.Second,
		},
		VotingParams: VotingParams{
			VotingPeriod: time.Duration(172800) * time.Second,
		},
		TallyParams: TallyParams{
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
	k.setDepositParams(ctx, data.DepositParams)
	k.setVotingParams(ctx, data.VotingParams)
	k.setTallyParams(ctx, data.TallyParams)
	for _, deposit := range data.Deposits {
		k.setDeposit(ctx, deposit.ProposalID, deposit.Deposit.Depositor, deposit.Deposit)
	}
	for _, vote := range data.Votes {
		k.setVote(ctx, vote.ProposalID, vote.Vote.Voter, vote.Vote)
	}
	for _, proposal := range data.Proposals {
		k.SetProposal(ctx, proposal)
	}
}

// ExportGenesis - output genesis parameters
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	startingProposalID, _ := k.peekCurrentProposalID(ctx)
	depositParams := k.GetDepositParams(ctx)
	votingParams := k.GetVotingParams(ctx)
	tallyParams := k.GetTallyParams(ctx)
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
		DepositParams:      depositParams,
		VotingParams:       votingParams,
		TallyParams:        tallyParams,
	}
}
