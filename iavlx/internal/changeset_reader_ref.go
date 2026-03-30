package internal

import "sync/atomic"

// ChangesetReaderRef wraps a ChangesetReader with reference counting to manage its lifecycle.
//
// The problem this solves: a ChangesetReader mmaps data files on disk. When a new checkpoint is
// written, we need to swap in a fresh reader (with updated mmaps covering the new data). But
// existing callers may still be using the old reader — we can't close its mmaps out from under them.
//
// The solution is pin/unpin reference counting:
//   - TryPin increments the refcount and returns the reader. If the ref has been evicted
//     (a new reader replaced it), TryPin returns nil.
//   - When the caller is done, they Unpin (via Pin.Unpin), decrementing the refcount.
//   - When a new reader replaces this one, Evict is called, which marks it as evicted and
//     queues it for disposal in the cleanup proc.
//   - The cleanup proc periodically calls TryDispose, which closes the underlying mmaps
//     once the refcount drops to zero.
//
// This ensures mmaps are never closed while someone is reading from them, even during
// concurrent checkpoint writes and compaction.
type ChangesetReaderRef struct {
	rdr       *ChangesetReader
	refCount  atomic.Int32
	evicted   atomic.Bool
	disposed  atomic.Bool
	changeset *Changeset
}

func (p *ChangesetReaderRef) Evict() {
	p.evicted.Store(true)
	p.rdr.changeset.treeStore.addToDisposalQueue(p)
}

// TryPin attempts to pin the ChangesetReader.
// If it is evicted, it returns (nil, Pin{}).
func (p *ChangesetReaderRef) TryPin() (*ChangesetReader, Pin) {
	p.refCount.Add(1)
	if p.evicted.Load() {
		p.refCount.Add(-1)
		return nil, Pin{}
	}
	return p.rdr, newPin(p)
}

func (p *ChangesetReaderRef) TryDispose() (bool, error) {
	if p.disposed.Load() {
		return true, nil
	}
	// TODO do we need to check evicted here?
	if p.refCount.Load() > 0 {
		return false, nil
	}
	p.disposed.Store(true)
	p.changeset.activeReaderCount.Add(-1)
	return true, p.rdr.Close()
}
