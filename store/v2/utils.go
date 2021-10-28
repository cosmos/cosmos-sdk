package types

import (
	"fmt"
	v1 "github.com/cosmos/cosmos-sdk/store/types"
)

var PrefixEndBytes = v1.PrefixEndBytes

// // Iterator over all the keys with a certain prefix in ascending order
// func KVStorePrefixIterator(kvs KVStore, prefix []byte) Iterator {
// 	return kvs.Iterator(prefix, PrefixEndBytes(prefix))
// }

// // Iterator over all the keys with a certain prefix in descending order.
// func KVStoreReversePrefixIterator(kvs KVStore, prefix []byte) Iterator {
// 	return kvs.ReverseIterator(prefix, PrefixEndBytes(prefix))
// }

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
