package gov

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
			MinDeposit:       sdk.Coins{sdk.NewInt64Coin(stakingTypes.DefaultBondDenom, 10)},
			MaxDepositPeriod: time.Duration(172800) * time.Second,
		},
		VotingParams: VotingParams{
			VotingPeriod: time.Duration(172800) * time.Second,
		},
		TallyParams: TallyParams{
			Quorum:            sdk.NewDecWithPrec(334, 3),
			Threshold:         sdk.NewDecWithPrec(5, 1),
			Veto:              sdk.NewDecWithPrec(334, 3),
			GovernancePenalty: sdk.NewDecWithPrec(1, 2),
		},
	}
}

// Checks whether 2 GenesisState structs are equivalent.
func (data GenesisState) Equal(data2 GenesisState) bool {
	if data.StartingProposalID != data.StartingProposalID ||
		!data.DepositParams.Equal(data2.DepositParams) ||
		data.VotingParams != data2.VotingParams ||
		data.TallyParams != data2.TallyParams {
		return false
	}

	if len(data.Deposits) != len(data2.Deposits) {
		return false
	}
	for i := range data.Deposits {
		deposit1 := data.Deposits[i]
		deposit2 := data2.Deposits[i]
		if deposit1.ProposalID != deposit2.ProposalID ||
			!deposit1.Deposit.Equals(deposit2.Deposit) {
			return false
		}
	}

	if len(data.Votes) != len(data2.Votes) {
		return false
	}
	for i := range data.Votes {
		vote1 := data.Votes[i]
		vote2 := data2.Votes[i]
		if vote1.ProposalID != vote2.ProposalID ||
			!vote1.Vote.Equals(vote2.Vote) {
			return false
		}
	}

	if len(data.Proposals) != len(data2.Proposals) {
		return false
	}
	for i := range data.Proposals {
		if data.Proposals[i] != data.Proposals[i] {
			return false
		}
	}

	return true

}

// Returns if a GenesisState is empty or has data in it
func (data GenesisState) IsEmpty() bool {
	emptyGenState := GenesisState{}
	return data.Equal(emptyGenState)
}

// ValidateGenesis TODO https://github.com/cosmos/cosmos-sdk/issues/3007
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

	govPenalty := data.TallyParams.GovernancePenalty
	if govPenalty.IsNegative() || govPenalty.GT(sdk.OneDec()) {
		return fmt.Errorf("Governance vote veto threshold should be positive and less or equal to one, is %s",
			govPenalty.String())
	}

	if data.DepositParams.MaxDepositPeriod > data.VotingParams.VotingPeriod {
		return fmt.Errorf("Governance deposit period should be less than or equal to the voting period (%ds), is %ds",
			data.VotingParams.VotingPeriod, data.DepositParams.MaxDepositPeriod)
	}

	if !data.DepositParams.MinDeposit.IsValid() {
		return fmt.Errorf("Governance deposit amount must be a valid sdk.Coins amount, is %s",
			data.DepositParams.MinDeposit.String())
	}

	return nil
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
		switch proposal.GetStatus() {
		case StatusDepositPeriod:
			k.InsertInactiveProposalQueue(ctx, proposal.GetDepositEndTime(), proposal.GetProposalID())
		case StatusVotingPeriod:
			k.InsertActiveProposalQueue(ctx, proposal.GetVotingEndTime(), proposal.GetProposalID())
		}
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
