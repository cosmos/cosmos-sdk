package keeper

const (
	// ModuleName is the name of the module
	ModuleName = "supply"

	// StoreKey is the default store key for supply
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the supply store.
	QuerierRoute = StoreKey
)

var supplierKey = []byte{0x0}