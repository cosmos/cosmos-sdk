package cometbft

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMaximumBlockGas gets the maximum gas from the consensus params. It panics
// if maximum block gas is less than negative one and returns zero if negative
// one.
func (w *cometABCIWrapper) GetMaximumBlockGas(ctx sdk.Context) uint64 {
	cp := w.GetConsensusParams(ctx)

	if cp.Block == nil {
		return 0
	}

	maxGas := cp.Block.MaxGas

	switch {
	case maxGas < -1:
		panic(fmt.Sprintf("invalid maximum block gas: %d", maxGas))

	case maxGas == -1:
		return 0

	default:
		return uint64(maxGas)
	}
}
