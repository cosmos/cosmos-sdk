package internal

// Pin represents a handle that pins some memory-mapped file data in memory.
// When the Pin is released via Unpin(), the data may be unmapped from memory.
// Pin must be used to ensure that any UnsafeBytes obtained from memory-mapped
// data remains valid while in use.
//
// It is recommended to use the following pattern:
//
//	node, pin, err := nodePointer.Resolve()
//	defer pin.Unpin()
//	if err != nil {
//		// handle error
//	}
//
// The caller must ensure that Unpin() is called at least once for each Pin obtained.
// Calling Unpin more than once is guaranteed to be idempotent.
// However, two threads should never try to Unpin the same Pin concurrently.
// The zero value of Pin is a no-op Pin; calling Unpin on a zero-value Pin has no effect and is always a safe return value.
// The convention is to pass pins by value so that there can be no nil-pointer dereference issues.
//
// Because calls to Unpin are idempotent, it is safe to do the following in a for loop:
//
//	for {
//		node, pin, err := nodePointer.Resolve()
//		defer pin.Unpin() // defer ensures that the pin will be released when the function returns early
//		if err != nil {
//			return err // early return
//		}
//	 // do something with node
//	 pin.Unpin() // can safely unpin here too
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
type Pin struct {
	// it is customary to pass Pin by value, so we use a pointer to an unpinner struct
	// to ensure that the unpinning is only done once even if multiple copies of
	// the Pin are passed around
	unpinner *unpinner
}

// newPin creates a new Pin for the given ChangesetReaderRef
// which will decrement the ref count of the ChangesetReaderRef when Unpin is called.
func newPin(ref *ChangesetReaderRef) Pin {
	return Pin{unpinner: &unpinner{ref}}
}

// unpinner is a simple struct that holds the ChangesetReaderRef to unpin from
// for now this type is specialized around decrementing the ref count of a ChangesetReaderRef
// but in the future it could be extended to support other types of pins if needed either with an interface or closure
// for now this optimization reducing heap alloc and CPU cycles a bit
type unpinner struct {
	ref *ChangesetReaderRef
}

// Unpin releases the Pin, allowing the underlying memory to be unmapped.
// Calling Unpin more than once is guaranteed to be safe and idempotent,
// but concurrent calls to Unpin on the same Pin should be avoided.
func (p *Pin) Unpin() {
	// this design ensures that when Pin is passed around by value - which is the convention -
	// we only unpin the memory once when Unpin is called, and subsequent calls to Unpin will be no-ops
	unpinner := p.unpinner
	if unpinner == nil {
		return
	}
	ref := unpinner.ref
	if ref == nil {
		return
	}
	ref.refCount.Add(-1)
	unpinner.ref = nil
	p.unpinner = nil
}
