package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName is the name of the module
	ModuleName = "supply"

	// StoreKey is the default store key for supply
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the supply store.
	QuerierRoute = StoreKey
)

// DefaultCodespace from the supply module
var DefaultCodespace sdk.CodespaceType = ModuleName

var supplyKey = []byte{0x0}
