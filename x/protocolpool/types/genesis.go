package types

import (
	"errors"

	"cosmossdk.io/math"
)

func NewGenesisState(cf []ContinuousFund) *GenesisState {
	return &GenesisState{
		ContinuousFunds: cf,
		Params:          Params{},
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ContinuousFunds: []ContinuousFund{},
		Params:          DefaultParams(),
	}
}

// ValidateGenesis validates the genesis state of protocolpool genesis input
func ValidateGenesis(gs *GenesisState) error {
	totalPercentage := math.LegacyZeroDec()
	for _, cf := range gs.ContinuousFunds {
		totalPercentage = totalPercentage.Add(cf.Percentage)
		if err := cf.Validate(); err != nil {
			return err
		}
	}
	if totalPercentage.GT(math.LegacyOneDec()) {
		return errors.New("total percentage cannot be greater than 100")
	}

	return gs.Params.Validate()
}
