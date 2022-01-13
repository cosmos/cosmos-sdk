package memdb

import (
	"bytes"
	"context"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/google/btree"
)

const (
	// Size of the channel buffer between traversal goroutine and iterator. Using an unbuffered
	// channel causes two context switches per item sent, while buffering allows more work per
	// context switch. Tuned with benchmarks.
	chBufferSize = 64
)

// memDBIterator is a memDB iterator.
type memDBIterator struct {
	ch     <-chan *item
	cancel context.CancelFunc
	item   *item
	start  []byte
	end    []byte
}

var _ db.Iterator = (*memDBIterator)(nil)

// newMemDBIterator creates a new memDBIterator.
// A visitor is passed to the btree which streams items to the iterator over a channel. Advancing
// the iterator pulls items from the channel, returning execution to the visitor.
// The reverse case needs some special handling, since we use [start, end) while btree uses (start, end]
func newMemDBIterator(tx *dbTxn, start []byte, end []byte, reverse bool) *memDBIterator {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *item, chBufferSize)
	iter := &memDBIterator{
		ch:     ch,
		cancel: cancel,
		start:  start,
		end:    end,
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
			item := i.(*item)
			if skipEqual != nil && bytes.Equal(item.key, skipEqual) {
				skipEqual = nil
				return true
			}
			if abortLessThan != nil && bytes.Compare(item.key, abortLessThan) == -1 {
				return false
			}
			select {
			case <-ctx.Done():
				return false
			case ch <- item:
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
	i.item = nil
	return nil
}

// Domain implements Iterator.
func (i *memDBIterator) Domain() ([]byte, []byte) {
	return i.start, i.end
}

// Next implements Iterator.
func (i *memDBIterator) Next() bool {
	item, ok := <-i.ch
	switch {
	case ok:
		i.item = item
	default:
		i.item = nil
	}
	return i.item != nil
}

// Error implements Iterator.
func (i *memDBIterator) Error() error {
	return nil
}

// Key implements Iterator.
func (i *memDBIterator) Key() []byte {
	i.assertIsValid()
	return i.item.key
}

// Value implements Iterator.
func (i *memDBIterator) Value() []byte {
	i.assertIsValid()
	return i.item.value
}

func (i *memDBIterator) assertIsValid() {
	if i.item == nil {
		panic("iterator is invalid")
	}
}
