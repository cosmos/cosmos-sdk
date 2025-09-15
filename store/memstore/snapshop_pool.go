package memstore

import (
	"sync"

	"cosmossdk.io/store/types"
)

type (
	snapshotPool struct {
		limit int64
		list  []*snapshotItem
	}

	SnapshotPool interface {
		Get(height int64) (types.MemStoreManager, bool)

		Set(height int64, store types.MemStoreManager)

		Limit(length int64)
	}

	snapshotItem struct {
		mtx    *sync.RWMutex
		store  types.MemStoreManager
		height int64
	}
)

const defaultLimit = 10

func newSnapshotPool() *snapshotPool {
	list := make([]*snapshotItem, defaultLimit)
	for i := 0; i < defaultLimit; i++ {
		list[i] = &snapshotItem{
			mtx:    &sync.RWMutex{},
			store:  nil,
			height: 0,
		}
	}

	return &snapshotPool{defaultLimit, list}
}

func (p *snapshotPool) Get(height int64) (types.MemStoreManager, bool) {
	idx := height % p.limit

	p.list[idx].mtx.RLock()
	defer p.list[idx].mtx.RUnlock()

	item := p.list[idx]
	if item.height != height {
		return nil, false
	}

	return item.store, item.store != nil
}

func (p *snapshotPool) Set(height int64, store types.MemStoreManager) {
	idx := height % p.limit

	p.list[idx].mtx.Lock()
	p.list[idx].store = store
	p.list[idx].height = height
	p.list[idx].mtx.Unlock()
}

func (p *snapshotPool) Limit(limit int64) {
	p.limit = limit
	p.list = make([]*snapshotItem, limit)
	for i := int64(0); i < limit; i++ {
		p.list[i] = &snapshotItem{
			mtx:    &sync.RWMutex{},
			store:  nil,
			height: 0,
		}
	}
}
