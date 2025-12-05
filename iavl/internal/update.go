package internal

// KVUpdate represents either a set or delete operation for a key-value pair.
// If SetNode is non-nil, it indicates a set operation.
// If DeleteKey is non-nil, it indicates a delete operation.
type KVUpdate struct {
	// SetNode uses a MemNode to represent a set operation.
	// If non-nil, it indicates that the key-value pair should be set/updated.
	// We use a *MemNode directly here because when KV data is serialized, we track
	// the KV offset inside the MemNode.kvOffset field which is later used
	// to reference the KV data in serialized tree nodes.
	SetNode *MemNode

	// DeleteKey uses a byte slice to represent the key being deleted.
	DeleteKey []byte
}
