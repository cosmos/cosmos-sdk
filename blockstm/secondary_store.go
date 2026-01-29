package blockstm

import (
	"sync"

	"github.com/RoaringBitmap/roaring/v2"
)

type secondaryDataItem[V any] struct {
	Incarnation Incarnation
	Value       V
	Estimate    bool
}

type SecondaryStore[V any] struct {
	// mu only protects bitmap, we don't sychonize between bitmap and data,
	// the reader can observe a version in bitmap but not in data, which is handled at reader side.
	mu     sync.RWMutex
	bitmap roaring.Bitmap

	data sync.Map
}

func NewSecondaryStore[V any]() *SecondaryStore[V] {
	return &SecondaryStore[V]{}
}

func (s *SecondaryStore[V]) Set(version TxnIndex, item secondaryDataItem[V]) {
	s.mu.Lock()
	s.bitmap.Add(uint32(version))
	s.mu.Unlock()

	s.data.Store(version, item)
}

func (s *SecondaryStore[V]) Delete(version TxnIndex) {
	s.mu.Lock()
	s.bitmap.Remove(uint32(version))
	s.mu.Unlock()

	s.data.Delete(version)
}

// PreviousValue returns the closest version that's less than the given version, exclusive.
func (s *SecondaryStore[V]) PreviousValue(target TxnIndex) (TxnIndex, secondaryDataItem[V], bool) {
	for target > 0 {
		s.mu.RLock()
		prev := s.bitmap.PreviousValue(uint32(target - 1))
		s.mu.RUnlock()

		if prev == -1 {
			return 0, secondaryDataItem[V]{}, false
		}

		target = TxnIndex(prev)
		value, ok := s.data.Load(target)
		if ok {
			return target, value.(secondaryDataItem[V]), true
		}
	}

	return 0, secondaryDataItem[V]{}, false
}

// Max is only called at block commit time, no need to synchronize.
func (s *SecondaryStore[V]) Max() (secondaryDataItem[V], bool) {
	if s.bitmap.IsEmpty() {
		return secondaryDataItem[V]{}, false
	}

	target := TxnIndex(s.bitmap.Maximum())

	value, ok := s.data.Load(target)
	if !ok {
		return secondaryDataItem[V]{}, false
	}

	return value.(secondaryDataItem[V]), true
}
