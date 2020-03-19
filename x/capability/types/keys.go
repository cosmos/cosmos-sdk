package types

import (
	"encoding/binary"
	"fmt"
)

const (
	// ModuleName defines the module name
	ModuleName = "capability"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"
)

var (
	// KeyIndex defines the key that stores the current globally unique capability
	// index.
	KeyIndex = []byte("index")

	// KeyPrefixIndexCapability defines a key prefix that stores index to capability
	// name mappings.
	KeyPrefixIndexCapability = []byte("capability_index")
)

// RevCapabilityKey returns a reverse lookup key for a given module and capability
// name.
func RevCapabilityKey(module, name string) []byte {
	return []byte(fmt.Sprintf("%s/rev/%s", module, name))
}

// FwdCapabilityKey returns a forward lookup key for a given module and capability
// reference.
func FwdCapabilityKey(module string, cap Capability) []byte {
	return []byte(fmt.Sprintf("%s/fwd/%s", module, cap))
}

// IndexToKey returns bytes to be used as a key for a given capability index.
func IndexToKey(index uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, index)
	return buf
}

// IndexFromKey returns an index from a call to IndexToKey for a given capability
// index.
func IndexFromKey(key []byte) uint64 {
	return binary.LittleEndian.Uint64(key)
}
