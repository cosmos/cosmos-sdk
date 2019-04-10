package supply

const (
	// ModuleName is the name of the module
	ModuleName = "supply"

	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the suply store.
	QuerierRoute = StoreKey
)

var (
	holderKeyPrefix = []byte{0x00}
)

// GetTokenHolderKey returns the store key of the given module
func GetTokenHolderKey(moduleName string) []byte {
	return append(holderKeyPrefix, []byte(moduleName)...)
}
