package blockstm

import (
	"sync"

	"github.com/RoaringBitmap/roaring/v2"
)

type BitmapIndex struct {
	sync.RWMutex
	bitmap roaring.Bitmap
}

func NewBitmapIndex() *BitmapIndex {
	return &BitmapIndex{}
}

func (s *BitmapIndex) Set(version TxnIndex) {
	s.Lock()
	s.bitmap.Add(uint32(version))
	s.Unlock()
}

func (s *BitmapIndex) Delete(version TxnIndex) {
	s.Lock()
	s.bitmap.Remove(uint32(version))
	s.Unlock()
}

// PreviousValue returns the closest version that's less than the given version, exclusive.
func (s *BitmapIndex) PreviousValue(target TxnIndex) (TxnIndex, bool) {
	s.RLock()
	prev := s.bitmap.PreviousValue(uint32(target - 1))
	s.RUnlock()

	if prev == -1 {
		return 0, false
	}

	return TxnIndex(prev), true
}

// Max is only called at block commit time, no need to synchronize.
func (s *BitmapIndex) Max() (TxnIndex, bool) {
	if s.bitmap.IsEmpty() {
		return 0, false
	}
	return TxnIndex(s.bitmap.Maximum()), true
}
