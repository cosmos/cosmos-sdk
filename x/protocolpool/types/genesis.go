package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

func NewGenesisState(cf []ContinuousFund) *GenesisState {
	return &GenesisState{
		ContinuousFund: cf,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ContinuousFund: []ContinuousFund{},
	}
}

// ValidateGenesis validates the genesis state of protocolpool genesis input
func ValidateGenesis(gs *GenesisState) error {
	for _, cf := range gs.ContinuousFund {
		err := validateContinuousFund(cf)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateContinuousFund(cf ContinuousFund) error {
	if cf.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if cf.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	if cf.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}

	// Validate percentage
	if cf.Percentage.IsZero() || cf.Percentage.IsNil() {
		return fmt.Errorf("percentage cannot be zero or empty")
	}
	if cf.Percentage.IsNegative() {
		return fmt.Errorf("percentage cannot be negative")
	}
	if cf.Percentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("percentage cannot be greater than one")
	}
	return nil
}
