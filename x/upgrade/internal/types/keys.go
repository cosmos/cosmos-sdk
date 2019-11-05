package types

const (
	// ModuleName is the name of this module
	ModuleName = "upgrade"

	// RouterKey is used to route governance proposals
	RouterKey = ModuleName

	// StoreKey is the prefix under which we store this module's data
	StoreKey = ModuleName

	// QuerierKey is used to handle abci_query requests
	QuerierKey = ModuleName
)

const (
	// PlanByte specifies the Byte under which a pending upgrade plan is stored in the store
	PlanByte = 0x0
	// DoneByte is a prefix for to look up completed upgrade plan by name
	DoneByte = 0x1
)

// PlanKey is the key under which the current plan is saved
// We store PlanByte as a const to keep it immutable (unlike a []byte)
func PlanKey() []byte {
	return []byte{PlanByte}
}
