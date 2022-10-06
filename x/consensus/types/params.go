package types

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

// Validate performs basic validation of ConsensusParams returning an error upon
// failure.
func Validate(p tmtypes.ConsensusParams) error {
	return p.ValidateBasic()
}
