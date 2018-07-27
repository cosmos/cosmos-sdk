package consensus

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	params "github.com/cosmos/cosmos-sdk/x/params/store"
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
func EndBlock(ctx sdk.Context, store params.Store) (updates *abci.ConsensusParams) {
	updates = &abci.ConsensusParams{
		BlockSize:   new(abci.BlockSize),
		TxSize:      new(abci.TxSize),
		BlockGossip: new(abci.BlockGossip),
	}

	if store.Modified(ctx, blockMaxBytesKey) {
		store.Get(ctx, blockMaxBytesKey, &updates.BlockSize.MaxBytes)
	}

	if store.Modified(ctx, blockMaxTxsKey) {
		store.Get(ctx, blockMaxTxsKey, &updates.BlockSize.MaxTxs)
	}

	if store.Modified(ctx, blockMaxGasKey) {
		store.Get(ctx, blockMaxGasKey, &updates.BlockSize.MaxGas)
	}

	if store.Modified(ctx, txMaxBytesKey) {
		store.Get(ctx, txMaxBytesKey, &updates.TxSize.MaxBytes)
	}

	if store.Modified(ctx, txMaxGasKey) {
		store.Get(ctx, txMaxGasKey, &updates.TxSize.MaxGas)
	}

	if store.Modified(ctx, blockPartSizeBytesKey) {
		store.Get(ctx, blockPartSizeBytesKey, &updates.BlockGossip.BlockPartSizeBytes)
	}

	return
}
