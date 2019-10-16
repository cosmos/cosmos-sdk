package types

const (
	// ModuleName is the name of the module
	ModuleName = "ibcmockbank"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the module
	RouterKey = ModuleName

	// codespace
	DefaultCodespace = ModuleName
)
