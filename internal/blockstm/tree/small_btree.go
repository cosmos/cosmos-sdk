package tree

import (
	"sync"
	"sync/atomic"
)

// SmallBTree is a small-optimized ordered set/map.
//
// It stores 0/1 item inline (lock-free reads) and promotes to a lock-based
// *BTree when a second distinct item is inserted.
type SmallBTree[T any] struct {
	mu     sync.Mutex
	less   func(a, b T) bool
	degree int

	slotA T
	slotB T

	// single points to either slotA or slotB.
	single atomic.Pointer[T]
	bt     atomic.Pointer[BTree[T]]
}

func NewSmallBTree[T any](less func(a, b T) bool, degree int) *SmallBTree[T] {
	return &SmallBTree[T]{
		less:   less,
		degree: degree,
	}
}

func (t *SmallBTree[T]) equal(a, b T) bool {
	return !t.less(a, b) && !t.less(b, a)
}

func (t *SmallBTree[T]) Get(item T) (result T, ok bool) {
	if bt := t.bt.Load(); bt != nil {
		return bt.Get(item)
	}
	if p := t.single.Load(); p != nil {
		if t.equal(*p, item) {
			return *p, true
		}
	}
	var zero T
	return zero, false
}

func (t *SmallBTree[T]) Set(item T) (prev T, ok bool) {
	if bt := t.bt.Load(); bt != nil {
		return bt.Set(item)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if bt := t.bt.Load(); bt != nil {
		return bt.Set(item)
	}

	if p := t.single.Load(); p != nil {
		cur := *p
		if t.equal(cur, item) {
			dst := &t.slotA
			if p == dst {
				dst = &t.slotB
			}
			*dst = item
			t.single.Store(dst)
			return cur, true
		}

		// Promote to a full B-tree.
		btNew := NewBTree(t.less, t.degree)
		btNew.Set(cur)
		btNew.Set(item)
		t.bt.Store(btNew)
		// Keep reads correct during promotion by storing bt first, then clearing single.
		t.single.Store(nil)
		var zero T
		return zero, false
	}

	dst := &t.slotA
	*dst = item
	t.single.Store(dst)
	var zero T
	return zero, false
}

func (t *SmallBTree[T]) Delete(item T) (prev T, ok bool) {
	if bt := t.bt.Load(); bt != nil {
		return bt.Delete(item)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if bt := t.bt.Load(); bt != nil {
		return bt.Delete(item)
	}

	if p := t.single.Load(); p != nil {
		cur := *p
		if t.equal(cur, item) {
			t.single.Store(nil)
			return cur, true
		}
	}
	var zero T
	return zero, false
}

// ReverseSeek returns the first item that is less than or equal to the pivot.
func (t *SmallBTree[T]) ReverseSeek(pivot T) (result T, ok bool) {
	if bt := t.bt.Load(); bt != nil {
		return bt.ReverseSeek(pivot)
	}
	if p := t.single.Load(); p != nil {
		// If pivot < item, there is no <= match.
		if t.less(pivot, *p) {
			var zero T
			return zero, false
		}
		return *p, true
	}
	var zero T
	return zero, false
}

func (t *SmallBTree[T]) Max() (result T, ok bool) {
	if bt := t.bt.Load(); bt != nil {
		return bt.Max()
	}
	if p := t.single.Load(); p != nil {
		return *p, true
	}
	var zero T
	return zero, false
}
