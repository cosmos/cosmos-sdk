package internal

import "sync/atomic"

type ChangesetReaderRef struct {
	rdr       *ChangesetReader
	refCount  atomic.Int32
	evicted   atomic.Bool
	disposed  atomic.Bool
	changeset *Changeset
}

type changesetReaderPin struct {
	pinner *ChangesetReaderRef
}

func (p *changesetReaderPin) Unpin() {
	if p.pinner == nil {
		return
	}
	p.pinner.refCount.Add(-1)
	p.pinner = nil
}

func (p *ChangesetReaderRef) Evict() {
	p.evicted.Store(true)
	p.rdr.changeset.treeStore.addToDisposalQueue(p)
}

// TryPin attempts to pin the ChangesetReader.
// If it is evicted, it returns (nil, NoopPin{}).
func (p *ChangesetReaderRef) TryPin() (*ChangesetReader, Pin) {
	p.refCount.Add(1)
	if p.evicted.Load() {
		p.refCount.Add(-1)
		return nil, NoopPin{}
	}
	return p.rdr, &changesetReaderPin{pinner: p}
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
