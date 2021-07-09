package v043

const (
	// ModuleName is the name of the module
	ModuleName = "bank"
)

// KVStore keys
var (
	BalancesPrefix = []byte{0x02}
	SupplyKey      = []byte{0x00}
)
