package types

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

// DefaultDefaultSendEnabled is the value that DefaultSendEnabled will have from DefaultParams().
var DefaultDefaultSendEnabled = true

// NewParams creates a new parameter configuration for the bank module
func NewParams(defaultSendEnabled bool) tmtypes.ConsensusParams {
	return tmtypes.ConsensusParams{}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() tmtypes.ConsensusParams {
	return tmtypes.ConsensusParams{}
}

// Validate all bank module parameters
func Validate(p tmtypes.ConsensusParams) error {
	return p.ValidateConsensusParams()

}

// String implements the Stringer interface.
func String(p tmtypes.ConsensusParams) string {
	cp := p.ToProto()
	return cp.String()
}
