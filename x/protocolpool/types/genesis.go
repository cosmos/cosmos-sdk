package types

import (
	"errors"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGenesisState(cf []ContinuousFund) *GenesisState {
	return &GenesisState{
		ContinuousFunds: cf,
		Params:          &Params{},
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ContinuousFunds: []ContinuousFund{},
		Params: &Params{
			EnabledDistributionDenoms: []string{sdk.DefaultBondDenom},
		},
	}
}

// ValidateGenesis validates the genesis state of protocolpool genesis input
func ValidateGenesis(gs *GenesisState) error {
	for _, cf := range gs.ContinuousFunds {
		if err := validateContinuousFund(cf); err != nil {
			return err
		}
	}
	return nil
}

func validateContinuousFund(cf ContinuousFund) error {
	if cf.Recipient == "" {
		return errors.New("recipient cannot be empty")
	}

	// Validate percentage
	if cf.Percentage.IsNil() || cf.Percentage.IsZero() {
		return errors.New("percentage cannot be zero or empty")
	}
	if cf.Percentage.IsNegative() {
		return errors.New("percentage cannot be negative")
	}
	if cf.Percentage.GT(math.LegacyOneDec()) {
		return errors.New("percentage cannot be greater than one")
	}
	return nil
}
