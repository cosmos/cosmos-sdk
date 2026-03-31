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
//
// When we are using arrays directly addressed to memory mapped files, these arrays
// are not part of the normal Go garbage collected memory. We must map and unmap
// these regions of memory explicitly. Pin represents a commitment to keep the memory
// mapped at least until Unpin() is called. During normal operation, changeset files
// will be mapped and unmapped as needed either because the file size has grown, we have
// compacted a changeset, or simply to manage open file descriptors.
// Under the hood pins use a reference counting mechanism to keep track of how many
// active users there are of a particular memory-mapped region.
type Pin interface {
	// Unpin releases the Pin, allowing the underlying memory to be unmapped.
	// Implementors should ensure that Unpin() is idempotent and only unpins the
	// memory once even if called multiple times.
	Unpin()
}

// NoopPin is a Pin that does nothing on Unpin().
type NoopPin struct{}

// Unpin implements the Pin interface.
func (NoopPin) Unpin() {}
