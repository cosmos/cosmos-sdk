package params

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keys for parameter access
const (
	BlockMaxBytesKey      = "params/BlockSize/MaxBytes"
	BlockMaxTxsKey        = "params/BlockSize/MaxTxs"
	BlockMaxGasKey        = "params/BlockSize/MaxGas"
	TxMaxBytesKey         = "params/TxSize/MaxBytes"
	TxMaxGasKey           = "params/TxSize/MaxGas"
	BlockPartSizeBytesKey = "params/BlockGossip/PartSizeBytes"
)

// nolint
func EndBlocker(ctx sdk.Context, k Keeper) (updates *abci.ConsensusParams) {
	if k.modified(ctx, BlockMaxBytesKey) {
		updates = new(abci.ConsensusParams)
		updates.BlockSize = new(abci.BlockSize)
		k.get(ctx, BlockMaxBytesKey, &updates.BlockSize.MaxBytes)
	}

	if k.modified(ctx, BlockMaxTxsKey) {
		if updates == nil {
			updates = new(abci.ConsensusParams)
		}
		if updates.BlockSize == nil {
			updates.BlockSize = new(abci.BlockSize)
		}
		k.get(ctx, BlockMaxTxsKey, &updates.BlockSize.MaxTxs)
	}

	if k.modified(ctx, BlockMaxGasKey) {
		if updates == nil {
			updates = new(abci.ConsensusParams)
		}
		if updates.BlockSize == nil {
			updates.BlockSize = new(abci.BlockSize)
		}
		k.get(ctx, BlockMaxGasKey, &updates.BlockSize.MaxTxs)
	}

	if k.modified(ctx, TxMaxBytesKey) {
		if updates == nil {
			updates = new(abci.ConsensusParams)
		}
		updates.TxSize = new(abci.TxSize)
		k.get(ctx, TxMaxBytesKey, &updates.BlockSize.MaxTxs)
	}

	if k.modified(ctx, TxMaxGasKey) {
		if updates == nil {
			updates = new(abci.ConsensusParams)
		}
		if updates.TxSize == nil {
			updates.TxSize = new(abci.TxSize)
		}
		k.get(ctx, TxMaxGasKey, &updates.BlockSize.MaxTxs)
	}

	return
}
