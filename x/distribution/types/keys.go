package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "distribution"

	// StoreKey is the store key string for distribution
	StoreKey = ModuleName

	// TStoreKey is the transient store key for distribution
	TStoreKey = "transient_" + ModuleName

	// RouterKey is the message route for distribution
	RouterKey = ModuleName

	// QuerierRoute is the querier route for distribution
	QuerierRoute = ModuleName
)
