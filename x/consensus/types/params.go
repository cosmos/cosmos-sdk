package types

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

// Validate all  module parameters
func Validate(p tmtypes.ConsensusParams) error {
	return p.ValidateBasic()
}
