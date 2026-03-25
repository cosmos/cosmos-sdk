package internal

// NodeUpdate represents either a set or delete operation for a key-value pair where sets are represented as new leaf
// MemNode's to be inserted into the tree.
// If SetNode is non-nil, it indicates a set operation.
// If DeleteKey is non-nil, it indicates a delete operation.
type NodeUpdate struct {
	// SetNode uses a MemNode to represent a set operation.
	// If non-nil, it indicates that the key-value pair should be set/updated.
	// We use a *MemNode directly here because when KV data is serialized, we track
	// the KV offset inside the MemNode.kvOffset field which is later used
	// to reference the KV data in serialized tree nodes.
	SetNode *MemNode

	// DeleteKey uses a byte slice to represent the key being deleted.
	DeleteKey []byte
}

// KVUpdate represents a key-value pair set or delete.
type KVUpdate = struct {
	// Key is the key to be set or deleted.
	Key []byte
	// Value is the value to be set. This should be nil when Delete is true.
	Value []byte
	// Delete indicates whether this update is a delete operation. If true, the key will be deleted.
	Delete bool
}
