package v1

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// NewGenesisState creates a new genesis state for the governance module
func NewGenesisState(startingProposalID uint64, dp DepositParams, vp VotingParams, tp TallyParams) *GenesisState {
	return &GenesisState{
		StartingProposalId: startingProposalID,
		DepositParams:      &dp,
		VotingParams:       &vp,
		TallyParams:        &tp,
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

// Empty returns true if a GenesisState is empty
func (data GenesisState) Empty() bool {
	return data.StartingProposalId == 0 ||
		data.DepositParams == nil ||
		data.VotingParams == nil ||
		data.TallyParams == nil
}

// ValidateGenesis checks if parameters are within valid ranges
func ValidateGenesis(data *GenesisState) error {
	if data.StartingProposalId == 0 {
		return errors.New("starting proposal id must be greater than 0")
	}

	if err := validateTallyParams(*data.TallyParams); err != nil {
		return fmt.Errorf("invalid tally params: %w", err)
	}

	if err := validateVotingParams(*data.VotingParams); err != nil {
		return fmt.Errorf("invalid voting params: %w", err)
	}

	if err := validateDepositParams(*data.DepositParams); err != nil {
		return fmt.Errorf("invalid deposit params: %w", err)
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
