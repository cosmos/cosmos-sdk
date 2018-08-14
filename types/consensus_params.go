package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Reusing tmtype
type ConsensusParams tmtypes.ConsensusParams

func (params ConsensusParams) ToABCI() *abci.ConsensusParams {
	inner := tmtypes.ConsensusParams(params)
	return tmtypes.TM2PB.ConsensusParams(&inner)
}

func (params *ConsensusParams) FromABCI(abciparams *abci.ConsensusParams) {
	// Manually set nil members to empty value
	if abciparams == nil {
		abciparams = &abci.ConsensusParams{
			BlockSize:   &abci.BlockSize{},
			TxSize:      &abci.TxSize{},
			BlockGossip: &abci.BlockGossip{},
		}
	} else {
		if abciparams.BlockSize == nil {
			abciparams.BlockSize = &abci.BlockSize{}
		}
		if abciparams.TxSize == nil {
			abciparams.TxSize = &abci.TxSize{}
		}
		if abciparams.BlockGossip == nil {
			abciparams.BlockGossip = &abci.BlockGossip{}
		}
	}

	*params = ConsensusParams(tmtypes.PB2TM.ConsensusParams(abciparams))
}
