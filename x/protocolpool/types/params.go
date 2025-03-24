package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultParams() Params {
	return Params{
		EnabledDistributionDenoms: []string{sdk.DefaultBondDenom},
		DistributionFrequency:     1,
	}
}

func (params *Params) Validate() error {
	for _, d := range params.EnabledDistributionDenoms {
		err := sdk.ValidateDenom(d)
		if err != nil {
			return err
		}
	}

	if params.DistributionFrequency == 0 {
		return errors.New("DistributionFrequency must be greater than 0")
	}

	return nil
}
