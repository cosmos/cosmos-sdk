package types

const (
	// ModuleName defines the module name.
	ModuleName = "epochs"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key.
	QuerierRoute = ModuleName
)

// KeyPrefixEpoch defines prefix key for storing epochs.
var KeyPrefixEpoch = []byte{0x01}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
