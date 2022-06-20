package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new genesis state for the governance module
func NewGenesisState(startingProposalID uint64, dp DepositParams, vp VotingParams, tp TallyParams) *GenesisState {
	return &GenesisState{
		StartingProposalId: startingProposalID,
		DepositParams:      dp,
		VotingParams:       vp,
		TallyParams:        tp,
	}
}

// DefaultGenesisState defines the default governance genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultStartingProposalID,
		DefaultDepositParams(),
		DefaultVotingParams(),
		DefaultTallyParams(),
	)
}

func (data GenesisState) Equal(other GenesisState) bool {
	return data.StartingProposalId == other.StartingProposalId &&
		data.Deposits.Equal(other.Deposits) &&
		data.Votes.Equal(other.Votes) &&
		data.Proposals.Equal(other.Proposals) &&
		data.DepositParams.Equal(other.DepositParams) &&
		data.TallyParams.Equal(other.TallyParams) &&
		data.VotingParams.Equal(other.VotingParams)
}

// Empty returns true if a GenesisState is empty
func (data GenesisState) Empty() bool {
	return data.Equal(GenesisState{})
}

// ValidateGenesis checks if parameters are within valid ranges
func ValidateGenesis(data *GenesisState) error {
	threshold := data.TallyParams.Threshold
	if threshold.IsNegative() || threshold.GT(sdk.OneDec()) {
		return fmt.Errorf("governance vote threshold should be positive and less or equal to one, is %s",
			threshold.String())
	}

	expeditedThreshold := data.TallyParams.ExpeditedThreshold
	if expeditedThreshold.IsNegative() || expeditedThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("governance expedited vote threshold should be positive and less or equal to one, is %s",
			expeditedThreshold)
	}

	if expeditedThreshold.LTE(threshold) {
		return fmt.Errorf("expedited governance vote threshold %s should be greater than the regular threshold %s",
			expeditedThreshold,
			threshold)
	}

	veto := data.TallyParams.VetoThreshold
	if veto.IsNegative() || veto.GT(sdk.OneDec()) {
		return fmt.Errorf("governance vote veto threshold should be positive and less or equal to one, is %s",
			veto.String())
	}

	minDeposit := data.DepositParams.MinDeposit
	if !minDeposit.IsValid() {
		return fmt.Errorf("governance deposit amount must be a valid sdk.Coins amount, is %s",
			minDeposit.String())
	}

	minExpeditedDeposit := data.DepositParams.MinExpeditedDeposit
	if !minExpeditedDeposit.IsValid() {
		return fmt.Errorf("governance expedited deposit amount must be a valid sdk.Coins amount, is %s",
			minExpeditedDeposit.String())
	}

	if minExpeditedDeposit.IsAllLTE(minDeposit) {
		return fmt.Errorf("governance min expedited deposit amount %s must be greater than regular min deposit %s",
			minExpeditedDeposit.String(),
			minDeposit.String())
	}

	return nil
}

var _ types.UnpackInterfacesMessage = GenesisState{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (data GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, p := range data.Proposals {
		err := p.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}
