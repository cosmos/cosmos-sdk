package types

const (
	// ModuleName is the name of the module.
	ModuleName = "uniswap"

	// RouterKey is the message route for the uniswap module.
	RouterKey = ModuleName

	// StoreKey is the default store key for the uniswap module.
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the uniswap module.
	QuerierRoute = StoreKey
)

// native asset to the module
type (
	NativeAsset string
)
