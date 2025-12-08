package internal

// Pin represents a handle that pins some memory-mapped file data in memory.
// When the Pin is released via Unpin(), the data may be unmapped from memory.
// Pin must be used to ensure that any UnsafeBytes obtained from memory-mapped
// data remains valid while in use.
// The caller must ensure that Unpin() is called exactly once
// for each Pin obtained. It is recommended to use the following pattern:
//
//	node, pin, err := nodePointer.Resolve()
//	defer pin.Unpin()
//	if err != nil {
//		// handle error
//	}
type Pin interface {
	// Unpin releases the Pin, allowing the underlying memory to be unmapped.
	Unpin()
}

// NoopPin is a Pin that has nothing on Unpin().
type NoopPin struct{}

// Unpin implements the Pin interface.
func (NoopPin) Unpin() {}
