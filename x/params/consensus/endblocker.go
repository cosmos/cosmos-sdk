package consensus

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	params "github.com/cosmos/cosmos-sdk/x/params/store"
)

// Keys for parameter access
const (
	DefaultParamSpace = "ConsensusParams"

	BlockSizeSpace   = "BlockSize"
	TxSizeSpace      = "TxSize"
	BlockGossipSpace = "BlockGossip"

	MaxBytesKey      = "MaxBytes"
	MaxTxsKey        = "MaxTxs"
	MaxGasKey        = "MaxGas"
	PartSizeBytesKey = "PartSizeBytes"
)

func BlockMaxBytesKey() params.Key      { return params.NewKey(BlockSizeSpace, MaxBytesKey) }
func BlockMaxTxsKey() params.Key        { return params.NewKey(BlockSizeSpace, MaxTxsKey) }
func BlockMaxGasKey() params.Key        { return params.NewKey(BlockSizeSpace, MaxGasKey) }
func TxMaxBytesKey() params.Key         { return params.NewKey(TxSizeSpace, MaxBytesKey) }
func TxMaxGasKey() params.Key           { return params.NewKey(TxSizeSpace, MaxGasKey) }
func BlockPartSizeBytesKey() params.Key { return params.NewKey(BlockGossipSpace, PartSizeBytesKey) }

var (
	blockMaxBytesKey      = BlockMaxBytesKey()
	blockMaxTxsKey        = BlockMaxTxsKey()
	blockMaxGasKey        = BlockMaxGasKey()
	txMaxBytesKey         = TxMaxBytesKey()
	txMaxGasKey           = TxMaxGasKey()
	blockPartSizeBytesKey = BlockPartSizeBytesKey()
)

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
