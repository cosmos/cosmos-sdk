package gov

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// Default period for deposits & voting
	DefaultPeriod time.Duration = 86400 * 2 * time.Second // 2 days
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	StartingProposalID uint64        `json:"starting_proposal_id" yaml:"starting_proposal_id"`
	Deposits           Deposits      `json:"deposits" yaml:"deposits"`
	Votes              Votes         `json:"votes" yaml:"votes"`
	Proposals          []Proposal    `json:"proposals" yaml:"proposals"`
	DepositParams      DepositParams `json:"deposit_params" yaml:"deposit_params"`
	VotingParams       VotingParams  `json:"voting_params" yaml:"voting_params"`
	TallyParams        TallyParams   `json:"tally_params" yaml:"tally_params"`
}

// NewGenesisState creates a new genesis state for the governance module
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
	minDepositTokens := sdk.TokensFromConsensusPower(10)
	return GenesisState{
		StartingProposalID: 1,
		DepositParams: DepositParams{
			MinDeposit:       sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, minDepositTokens)},
			MaxDepositPeriod: DefaultPeriod,
		},
		VotingParams: VotingParams{
			VotingPeriod: DefaultPeriod,
		},
		TallyParams: TallyParams{
			Quorum:    sdk.NewDecWithPrec(334, 3),
			Threshold: sdk.NewDecWithPrec(5, 1),
			Veto:      sdk.NewDecWithPrec(334, 3),
		},
	}
}

// Checks whether 2 GenesisState structs are equivalent.
func (data GenesisState) Equal(data2 GenesisState) bool {
	b1 := types.ModuleCdc.MustMarshalBinaryBare(data)
	b2 := types.ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// Returns if a GenesisState is empty or has data in it
func (data GenesisState) IsEmpty() bool {
	emptyGenState := GenesisState{}
	return data.Equal(emptyGenState)
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

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, supplyKeeper SupplyKeeper, data GenesisState) {

	k.setProposalID(ctx, data.StartingProposalID)
	k.setDepositParams(ctx, data.DepositParams)
	k.setVotingParams(ctx, data.VotingParams)
	k.setTallyParams(ctx, data.TallyParams)

	// check if the deposits pool account exists
	moduleAcc := k.GetGovernanceAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	var totalDeposits sdk.Coins
	for _, deposit := range data.Deposits {
		k.setDeposit(ctx, deposit.ProposalID, deposit.Depositor, deposit)
		totalDeposits = totalDeposits.Add(deposit.Amount)
	}

	for _, vote := range data.Votes {
		k.setVote(ctx, vote.ProposalID, vote.Voter, vote)
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
