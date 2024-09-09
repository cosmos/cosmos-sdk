package types

import (
	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "bankv2"

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	GovModuleName = "gov"

	// MintModuleName duplicates the mint module's name to avoid a cyclic dependency with x/mint.
	// It should be synced with the mint module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/0e34478eb7420b69869ed50f129fc274a97a9b06/x/mint/types/keys.go#L13
	MintModuleName = "mint"
)

var (
	// ParamsKey is the prefix for x/bank/v2 parameters
	ParamsKey = collections.NewPrefix(2)

	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix = collections.NewPrefix(3)

	DenomAddressPrefix = collections.NewPrefix(4)

	SupplyKey = collections.NewPrefix(5)
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
