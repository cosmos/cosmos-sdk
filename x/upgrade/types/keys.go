package types

import "fmt"

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

	// KeyUpgradedIBCState is the key under which upgraded ibc state is stored in the upgrade store
	KeyUpgradedIBCState = "upgradedIBCState"

	// KeyUpgradedClient is the sub-key under which upgraded client state will be stored
	KeyUpgradedClient = "upgradedClient"

	// KeyUpgradedConsState is the sub-key under which upgraded consensus state will be stored
	KeyUpgradedConsState = "upgradedConsState"
)

// PlanKey is the key under which the current plan is saved
// We store PlanByte as a const to keep it immutable (unlike a []byte)
func PlanKey() []byte {
	return []byte{PlanByte}
}

// UpgradedClientKey is the key under which the upgraded client state is saved
// Connecting IBC chains can verify against the upgraded client in this path before
// upgrading their clients
func UpgradedClientKey(height int64) []byte {
	return []byte(fmt.Sprintf("%s/%d/%s", KeyUpgradedIBCState, height, KeyUpgradedClient))
}

// UpgradedConsStateKey is the key under which the upgraded consensus state is saved
// Connecting IBC chains can verify against the upgraded consensus state in this path before
// upgrading their clients.
func UpgradedConsStateKey(height int64) []byte {
	return []byte(fmt.Sprintf("%s/%d/%s", KeyUpgradedIBCState, height, KeyUpgradedConsState))
}
