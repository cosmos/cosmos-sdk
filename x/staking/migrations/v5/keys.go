package v5

const (
	// ModuleName is the name of the module
	ModuleName = "staking"
)

var (
	DelegationKey           = []byte{0x31} // key for a delegation
	DelegationByValIndexKey = []byte{0x37} // key for delegations by a validator
)
