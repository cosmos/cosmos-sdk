package types

import (
	"fmt"
)

func StoreKeyToType(key StoreKey) (typ StoreType, err error) {
	switch key.(type) {
	case *KVStoreKey:
		typ = StoreTypePersistent
	case *MemoryStoreKey:
		typ = StoreTypeMemory
	case *TransientStoreKey:
		typ = StoreTypeTransient
	default:
		err = fmt.Errorf("unrecognized store key type: %T", key)
	}
	return
}
