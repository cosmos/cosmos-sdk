package types

import (
	"errors"
	"fmt"

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

// Validate validates the genesis state of protocolpool genesis input
func (gs *GenesisState) Validate() error {
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

	seenContinuousFunds := make(map[string]struct{})
	for _, fund := range gs.ContinuousFunds {
		if err := fund.Validate(); err != nil {
			return fmt.Errorf("invalid continuousfund: %w", err)
		}

		if _, ok := seenContinuousFunds[fund.Recipient]; ok {
			return fmt.Errorf("duplicated continuous fund recipient address: %s", fund.Recipient)
		}
		seenContinuousFunds[fund.Recipient] = struct{}{}
	}

	return gs.Params.Validate()
}
