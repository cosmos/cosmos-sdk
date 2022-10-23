package types

// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// OnWrite captures store state changes
	// if value is nil then it was deleted
	// storeKey indicates the source KVStore, to facilitate using the same WriteListener across separate KVStores
	// delete bool indicates if it was a delete; true: delete, false: set
	OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool)
}
