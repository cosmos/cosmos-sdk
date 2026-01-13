package internal

import "sync/atomic"

type ChangesetReaderRef struct {
	rdr      *ChangesetReader
	refCount atomic.Int32
	evicted  atomic.Bool
	disposed atomic.Bool
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
}

func (p *ChangesetReaderRef) TryPin() (*ChangesetReader, Pin) {
	p.refCount.Add(1)
	if p.evicted.Load() {
		p.refCount.Add(-1)
		return nil, nil
	}
	return p.rdr, &changesetReaderPin{pinner: p}
}

//func (p *ChangesetReaderRef) TryDispose() bool {
//	if p.disposed.Load() {
//		return true
//	}
//	// TODO do we need to check evicted here?
//	if p.refCount.Load() > 0 {
//		return false
//	}
//	p.disposed.Store(true)
//	_ = p.rdr.Close()
//	return true
//}
