package memdb

import (
	"bytes"
	"context"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/google/btree"
)

const (
	// TODO: this neeeds to be tuned or maybe even made configurable
	bufferSize = 1024
)

// memDBIterator is a memDB iterator.
type memDBIterator struct {
	ctx     context.Context
	cancel  context.CancelFunc
	start   []byte
	end     []byte
	reverse bool
	tx      *dbTxn

	// Because we use [start, end) for reverse ranges, while btree uses (start, end], we need
	// the following variables to handle some reverse iteration conditions ourselves.
	skipEqual     []byte
	abortLessThan []byte

	items  []*item
	cursor int
}

var _ db.Iterator = (*memDBIterator)(nil)

// newMemDBIterator creates a new memDBIterator.
// A visitor is passed to the btree which streams items to the iterator over a channel. Advancing
// the iterator pulls items from the channel, returning execution to the visitor.
// The reverse case needs some special handling, since we use [start, end) while btree uses (start, end]
func newMemDBIterator(tx *dbTxn, start []byte, end []byte, reverse bool) *memDBIterator {
	ctx, cancel := context.WithCancel(context.Background())
	iter := &memDBIterator{
		ctx:     ctx,
		cancel:  cancel,
		start:   start,
		end:     end,
		reverse: reverse,
		cursor:  -1,
		items:   make([]*item, 0, bufferSize),
		tx:      tx,
	}

	iter.loadMore()

	return iter
}

// Close implements Iterator.
func (i *memDBIterator) Close() error {
	i.cancel()
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
	if i.cursor < len(i.items) {
		return i.items[i.cursor] != nil
	}

	// If the cursor is already over the current items length and the length is less than the buffer size;
	// It means we've reached the end
	if len(i.items) < bufferSize {
		return false
	}

	i.start = i.items[len(i.items)-1].key
	i.cursor = 0
	// skip the first item, since it's already loaded
	i.skipEqual = i.start
	i.loadMore()

	return len(i.items) > 0 && i.items[i.cursor] != nil
}

// loadMore starts or continues the iteration
func (i *memDBIterator) loadMore() {
	i.items = make([]*item, 0, bufferSize)
	visitor := func(it btree.Item) bool {
		nitem := it.(*item)
		if i.skipEqual != nil && bytes.Equal(nitem.key, i.skipEqual) {
			i.skipEqual = nil
			return true
		}
		if i.abortLessThan != nil && bytes.Compare(nitem.key, i.abortLessThan) == -1 {
			return false
		}

		select {
		case <-i.ctx.Done():
			return false
		default:
			i.items = append(i.items, nitem)
			// once the buffer is full, stop the iteration
			if len(i.items) == bufferSize {
				return false
			}
			return true
		}
	}
	switch {
	case i.start == nil && i.end == nil && !i.reverse:
		i.tx.btree.Ascend(visitor)
	case i.start == nil && i.end == nil && i.reverse:
		i.tx.btree.Descend(visitor)
	case i.end == nil && !i.reverse:
		// must handle this specially, since nil is considered less than anything else
		i.tx.btree.AscendGreaterOrEqual(newKey(i.start), visitor)
	case !i.reverse:
		i.tx.btree.AscendRange(newKey(i.start), newKey(i.end), visitor)
	case i.end == nil:
		// abort after start, since we use [start, end) while btree uses (start, end]
		i.abortLessThan = i.start
		i.tx.btree.Descend(visitor)
	default:
		// skip end and abort after start, since we use [start, end) while btree uses (start, end]
		i.skipEqual = i.end
		i.abortLessThan = i.start
		i.tx.btree.DescendLessOrEqual(newKey(i.end), visitor)
	}

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
