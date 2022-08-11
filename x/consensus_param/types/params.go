package types

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

// Validate all  module parameters
func Validate(p tmtypes.ConsensusParams) error {
	return p.ValidateConsensusParams()

}

// String implements the Stringer interface.
func String(p tmtypes.ConsensusParams) string {
	cp := p.ToProto()
	return cp.String()
}
