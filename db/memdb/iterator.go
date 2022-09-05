package memdb

import (
	"bytes"
	"context"
	"sync"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/google/btree"
)

const (
	// Size of the channel buffer between traversal goroutine and iterator. Using an unbuffered
	// channel causes two context switches per item sent, while buffering allows more work per
	// context switch. Tuned with benchmarks.
	chBufferSize = 64
	sliceSize    = 64
)

// memDBIterator is a memDB iterator.
type memDBIterator struct {
	ch     <-chan []*item
	cancel context.CancelFunc
	start  []byte
	end    []byte

	items     []*item
	cursor    int
	buffer    []*item
	buffermtx sync.RWMutex
}

var _ db.Iterator = (*memDBIterator)(nil)

// newMemDBIterator creates a new memDBIterator.
// A visitor is passed to the btree which streams items to the iterator over a channel. Advancing
// the iterator pulls items from the channel, returning execution to the visitor.
// The reverse case needs some special handling, since we use [start, end) while btree uses (start, end]
func newMemDBIterator(tx *dbTxn, start []byte, end []byte, reverse bool) *memDBIterator {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan []*item, chBufferSize)
	iter := &memDBIterator{
		ch:     ch,
		cancel: cancel,
		start:  start,
		end:    end,
		cursor: -1,
		buffer: make([]*item, 0, sliceSize),
		items:  make([]*item, 0, sliceSize),
	}

	go func() {
		defer close(ch)
		// Because we use [start, end) for reverse ranges, while btree uses (start, end], we need
		// the following variables to handle some reverse iteration conditions ourselves.
		var (
			skipEqual     []byte
			abortLessThan []byte
		)
		visitor := func(i btree.Item) bool {
			nitem := i.(*item)
			if skipEqual != nil && bytes.Equal(nitem.key, skipEqual) {
				skipEqual = nil
				if len(iter.buffer) > 0 {
					iter.buffermtx.Lock()
					ch <- iter.buffer
					iter.buffer = make([]*item, 0, sliceSize)
					iter.buffermtx.Unlock()
				}
				return true
			}
			if abortLessThan != nil && bytes.Compare(nitem.key, abortLessThan) == -1 {
				if len(iter.buffer) > 0 {
					iter.buffermtx.Lock()
					ch <- iter.buffer
					iter.buffer = make([]*item, 0, sliceSize)
					iter.buffermtx.Unlock()
				}
				return false
			}

			iter.buffer = append(iter.buffer, nitem)
			if len(iter.buffer) == sliceSize {
				iter.buffermtx.Lock()
				ch <- iter.buffer
				iter.buffer = make([]*item, 0, sliceSize)
				iter.buffermtx.Unlock()
			}

			select {
			case <-ctx.Done():
				iter.buffermtx.Lock()
				ch <- iter.buffer
				iter.buffer = make([]*item, 0, sliceSize)
				iter.buffermtx.Unlock()
				return false
			default:
				return true
			}
		}
		switch {
		case start == nil && end == nil && !reverse:
			tx.btree.Ascend(visitor)
		case start == nil && end == nil && reverse:
			tx.btree.Descend(visitor)
		case end == nil && !reverse:
			// must handle this specially, since nil is considered less than anything else
			tx.btree.AscendGreaterOrEqual(newKey(start), visitor)
		case !reverse:
			tx.btree.AscendRange(newKey(start), newKey(end), visitor)
		case end == nil:
			// abort after start, since we use [start, end) while btree uses (start, end]
			abortLessThan = start
			tx.btree.Descend(visitor)
		default:
			// skip end and abort after start, since we use [start, end) while btree uses (start, end]
			skipEqual = end
			abortLessThan = start
			tx.btree.DescendLessOrEqual(newKey(end), visitor)
		}
	}()

	return iter
}

// Close implements Iterator.
func (i *memDBIterator) Close() error {
	i.cancel()
	for range i.ch { // drain channel
	}
	i.items = nil
	return nil
}

// Domain implements Iterator.
func (i *memDBIterator) Domain() ([]byte, []byte) {
	return i.start, i.end
}

// Next implements Iterator.
func (i *memDBIterator) Next() bool {
	i.cursor++
	if i.cursor < len(i.items) && i.cursor != sliceSize {
		return i.items[i.cursor] != nil
	}

	it, ok := <-i.ch
	switch {
	case ok:
		i.items = it
		i.cursor = 0
	case len(i.buffer) > 0:
		i.buffermtx.RLock()
		defer i.buffermtx.RUnlock()
		i.items = i.buffer
		i.buffer = nil
		if len(i.items) > 0 {
			i.cursor = 0
		}
		return i.items[i.cursor] != nil
	default:
		return false
	}
	return i.items[i.cursor] != nil
}

// Error implements Iterator.
func (i *memDBIterator) Error() error {
	return nil
}

// Key implements Iterator.
func (i *memDBIterator) Key() []byte {
	i.assertIsValid()
	return i.items[i.cursor].key
}

// Value implements Iterator.
func (i *memDBIterator) Value() []byte {
	i.assertIsValid()
	return i.items[i.cursor].value
}

func (i *memDBIterator) assertIsValid() {
	if i.items[i.cursor] == nil {
		panic("iterator is invalid")
	}
}
