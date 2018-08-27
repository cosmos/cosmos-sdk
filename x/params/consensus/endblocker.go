package consensus

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// Default parameter namespace
const (
	DefaultParamSpace = "ConsensusParams"
)

// nolint - Key generators for parameter access
func BlockMaxBytesKey() params.Key      { return params.NewKey("BlockSize", "MaxBytes") }
func BlockMaxTxsKey() params.Key        { return params.NewKey("BlockSize", "MaxTxs") }
func BlockMaxGasKey() params.Key        { return params.NewKey("BlockSize", "MaxGas") }
func TxMaxBytesKey() params.Key         { return params.NewKey("TxSize", "MaxBytes") }
func TxMaxGasKey() params.Key           { return params.NewKey("TxSize", "MaxGas") }
func BlockPartSizeBytesKey() params.Key { return params.NewKey("BlockGossip", "PartSizeBytes") }

// Cached parameter keys
var (
	blockMaxBytesKey      = BlockMaxBytesKey()
	blockMaxTxsKey        = BlockMaxTxsKey()
	blockMaxGasKey        = BlockMaxGasKey()
	txMaxBytesKey         = TxMaxBytesKey()
	txMaxGasKey           = TxMaxGasKey()
	blockPartSizeBytesKey = BlockPartSizeBytesKey()
)

// EndBlock returns consensus parameters set in the block
func EndBlock(ctx sdk.Context, space params.Space) (updates *abci.ConsensusParams) {
	updates = &abci.ConsensusParams{
		BlockSize:   new(abci.BlockSize),
		TxSize:      new(abci.TxSize),
		BlockGossip: new(abci.BlockGossip),
	}

	if space.Modified(ctx, blockMaxBytesKey) {
		space.Get(ctx, blockMaxBytesKey, &updates.BlockSize.MaxBytes)
	}

	if space.Modified(ctx, blockMaxTxsKey) {
		space.Get(ctx, blockMaxTxsKey, &updates.BlockSize.MaxTxs)
	}

	if space.Modified(ctx, blockMaxGasKey) {
		space.Get(ctx, blockMaxGasKey, &updates.BlockSize.MaxGas)
	}

	if space.Modified(ctx, txMaxBytesKey) {
		space.Get(ctx, txMaxBytesKey, &updates.TxSize.MaxBytes)
	}

	if space.Modified(ctx, txMaxGasKey) {
		space.Get(ctx, txMaxGasKey, &updates.TxSize.MaxGas)
	}

	if space.Modified(ctx, blockPartSizeBytesKey) {
		space.Get(ctx, blockPartSizeBytesKey, &updates.BlockGossip.BlockPartSizeBytes)
	}

	return
}
