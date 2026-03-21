package types

import (
	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "bank"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// ObjectStoreKey defines the store name for the object store
	ObjectStoreKey = "object:" + ModuleName
)

// KVStore keys
var (
	SupplyKey           = collections.NewPrefix(0)
	DenomMetadataPrefix = collections.NewPrefix(1)
	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix     = collections.NewPrefix(2)
	DenomAddressPrefix = collections.NewPrefix(3)
	// SendEnabledPrefix is the prefix for the SendDisabled flags for a Denom.
	SendEnabledPrefix = collections.NewPrefix(4)

	// ParamsKey is the prefix for x/bank parameters
	ParamsKey = collections.NewPrefix(5)
)

// ObjectStore keys for virtual operations
var (
	// VirtualBalancePrefix is the prefix for virtual balance changes in ObjectStore.
	// Key format: VirtualBalancePrefix + address + txIndex (8 bytes)
	VirtualBalancePrefix = []byte{0x00}

	// VirtualSupplyPrefix is the prefix for virtual supply changes in ObjectStore.
	// Key format: VirtualSupplyPrefix + denom + txIndex (8 bytes)
	VirtualSupplyPrefix = []byte{0x01}
)

// BalanceValueCodec is a codec for encoding bank balances in a backwards compatible way.
// Historically, balances were represented as Coin, now they're represented as a simple math.Int
var BalanceValueCodec = collcodec.NewAltValueCodec(sdk.IntValue, func(bytes []byte) (math.Int, error) {
	c := new(sdk.Coin)
	err := c.Unmarshal(bytes)
	if err != nil {
		return math.Int{}, err
	}
	return c.Amount, nil
})
