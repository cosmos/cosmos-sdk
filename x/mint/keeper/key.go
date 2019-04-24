// nolint
package keeper

const (
	// ModuleName is the name of the module
	ModuleName = "minting"

	// default paramspace for params keeper
	DefaultParamspace = "mint"

	// StoreKey is the default store key for mint
	StoreKey = "mint"

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey
)

var (
	minterKey = []byte{0x00} // the one key to use for the keeper store

	// params store for inflation params
	ParamStoreKeyParams = []byte("params")
)
