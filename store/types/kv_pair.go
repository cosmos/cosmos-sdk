package types

// StoreKVPair defines a key-value pair with additional metadata that is used
// to track writes to an underlying SC store.
type StoreKVPair struct {
	StoreKey string
	Delete   bool
	Key      []byte
	Value    []byte
}
