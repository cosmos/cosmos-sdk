package types

const (
	// ModuleName defines the module name
	ModuleName = "evidence"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KVStore key prefixes
var (
	KeyPrefixEvidence = []byte{0x00}
)

func EvidenceKey(hash []byte) (key []byte) {
	key = make([]byte, len(KeyPrefixEvidence)+len(hash))
	copy(key, KeyPrefixEvidence)
	copy(key[len(KeyPrefixEvidence):], hash)
	return
}
