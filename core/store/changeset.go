package store

// StateChanges represents a set of changes to the state of an actor in storage.
type StateChanges struct {
	Actor        []byte   // actor represents the space in storage where state is stored, previously this was called a "storekey"
	StateChanges []KVPair // StateChanges is a list of key-value pairs representing the changes to the state.
}

// KVPair represents a change in a key and value of state.
// Remove being true signals the key must be removed from state.
type KVPair struct {
	// Key defines the key being updated.
	Key []byte
	// Value defines the value associated with the updated key.
	Value []byte
	// Remove is true when the key must be removed from state.
	Remove bool
}
