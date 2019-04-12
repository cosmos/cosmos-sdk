package bank

const (
	// ModuleName is the name of the module
	ModuleName = "bank"

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the suply store.
	QuerierRoute = StoreKey
)

var (
	supplierKey     = []byte{0x00}
	holderKeyPrefix = []byte{0x01}
)

// GetTokenHolderKey returns the store key of the given module
func GetTokenHolderKey(moduleName string) []byte {
	return append(holderKeyPrefix, []byte(moduleName)...)
}
