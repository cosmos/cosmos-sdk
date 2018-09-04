package consensus

import (
	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// Default parameter namespace
const (
	DefaultParamSpace = "consensus"
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
