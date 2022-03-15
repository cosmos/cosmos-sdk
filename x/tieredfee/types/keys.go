package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of this module
	ModuleName = "tieredfee"

	// RouterKey is used to route governance proposals
	RouterKey = ModuleName

	// StoreKey is the prefix under which we store this module's data
	StoreKey = ModuleName

	// QuerierKey is used to handle abci_query requests
	QuerierKey = ModuleName
)

// KVStore keys
var (
	// BlockGasUsedKey is the prefix for the block gas used store.
	BlockGasUsedKey = []byte{0x00}
	// GasPriceKeyPrefix is the prefix for the gas price store, which stored current gas prices for all tiers.
	GasPriceKeyPrefix = []byte{0x01}
)

// GasPriceKey returns the key for gas price of a certain tier.
func GasPriceKey(tier uint32) []byte {
	bz := sdk.Uint64ToBigEndian(uint64(tier))
	return append(GasPriceKeyPrefix, bz...)
}
