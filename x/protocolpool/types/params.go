package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultParams() Params {
	return Params{
		EnabledDistributionDenoms: []string{sdk.DefaultBondDenom},
		DistributionFrequency:     1,
	}
}

func (p *Params) Validate() error {
	seenDenoms := make(map[string]struct{})

	for _, d := range p.EnabledDistributionDenoms {
		err := sdk.ValidateDenom(d)
		if err != nil {
			return err
		}

		if _, seen := seenDenoms[d]; seen {
			return fmt.Errorf("duplicate enabled distribution denom %s", d)
		}
		seenDenoms[d] = struct{}{}
	}

	if p.DistributionFrequency == 0 {
		return errors.New("DistributionFrequency must be greater than 0")
	}

	return nil
}
