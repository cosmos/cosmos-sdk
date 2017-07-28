package state

import "encoding/binary"

var (
	headKey = []byte("h")
	tailKey = []byte("t")
	dataKey = []byte("d")
)

// QueueHeadKey gives us the key for the height at head of the queue
func QueueHeadKey() []byte {
	return headKey
}

// QueueTailKey gives us the key for the height at tail of the queue
func QueueTailKey() []byte {
	return tailKey
}

// QueueItemKey gives us the key to look up one item by sequence
func QueueItemKey(i uint64) []byte {
	return makeKey(i)
}

// Queue allows us to fill up a range of the db, and grab from either end
type Queue struct {
	store KVStore
	head  uint64 // if Size() > 0, the first element is here
	tail  uint64 // this is the first empty slot to Push() to
}

// NewQueue will load or initialize a queue in this state-space
//
// Generally, you will want to stack.PrefixStore() the space first
func NewQueue(store KVStore) *Queue {
	q := &Queue{store: store}
	q.head = q.getCount(headKey)
	q.tail = q.getCount(tailKey)
	return q
}

// Tail returns the next slot that Push() will use
func (q *Queue) Tail() uint64 {
	return q.tail
}

// Size returns how many elements are in the queue
func (q *Queue) Size() int {
	return int(q.tail - q.head)
}

// Push adds an element to the tail of the queue and returns it's location
func (q *Queue) Push(value []byte) uint64 {
	key := makeKey(q.tail)
	q.store.Set(key, value)
	q.tail++
	q.setCount(tailKey, q.tail)
	return q.tail - 1
}

// Pop gets an element from the end of the queue
func (q *Queue) Pop() []byte {
	if q.Size() <= 0 {
		return nil
	}
	key := makeKey(q.head)
	value := q.store.Get(key)
	q.head++
	q.setCount(headKey, q.head)
	return value
}

// Item looks at any element in the queue, without modifying anything
func (q *Queue) Item(seq uint64) []byte {
	if seq >= q.tail || seq < q.head {
		return nil
	}
	return q.store.Get(makeKey(seq))
}

func (q *Queue) setCount(key []byte, val uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, val)
	q.store.Set(key, b)
}

func (q *Queue) getCount(key []byte) (val uint64) {
	b := q.store.Get(key)
	if b != nil {
		val = binary.BigEndian.Uint64(b)
	}
	return val
}

// makeKey returns the key for a data point
func makeKey(val uint64) []byte {
	b := make([]byte, 8+len(dataKey))
	copy(b, dataKey)
	binary.BigEndian.PutUint64(b[len(dataKey):], val)
	return b
}
