package iavlx

import "bytes"

type Iterator struct {
	// domain of iteration, end is exclusive
	start, end []byte
	ascending  bool
	zeroCopy   bool

	// cache the next key-value pair
	key, value []byte

	err   error
	valid bool

	stack []Node
}

func NewIterator(start, end []byte, ascending bool, rootPtr *NodePointer, zeroCopy bool) *Iterator {
	iter := &Iterator{
		start:     start,
		end:       end,
		ascending: ascending,
		valid:     true,
		zeroCopy:  zeroCopy,
	}

	if rootPtr != nil {
		root, err := rootPtr.Resolve()
		if err != nil {
			panic(err)
		}
		iter.stack = []Node{root}
	}

	// cache the first key-value
	iter.Next()
	return iter
}

func (iter *Iterator) Domain() ([]byte, []byte) {
	return iter.start, iter.end
}

// Valid implements dbm.Iterator.
func (iter *Iterator) Valid() bool {
	return iter.valid
}

// Error implements dbm.Iterator
func (iter *Iterator) Error() error {
	return nil
}

// Key implements dbm.Iterator
func (iter *Iterator) Key() []byte {
	if !iter.zeroCopy {
		return bytes.Clone(iter.key)
	}
	return iter.key
}

// Value implements dbm.Iterator
func (iter *Iterator) Value() []byte {
	if !iter.zeroCopy {
		return bytes.Clone(iter.value)
	}
	return iter.value
}

// Next implements dbm.Iterator
func (iter *Iterator) Next() {
	if !iter.valid {
		return
	}

	for len(iter.stack) > 0 {
		// pop node
		node := iter.stack[len(iter.stack)-1]
		iter.stack = iter.stack[:len(iter.stack)-1]

		key, err := node.Key()
		if err != nil {
			iter.fail(err)
			return
		}
		startCmp := bytes.Compare(iter.start, key)
		afterStart := iter.start == nil || startCmp < 0
		beforeEnd := iter.end == nil || bytes.Compare(key, iter.end) < 0

		if node.IsLeaf() {
			startOrAfter := afterStart || startCmp == 0
			if startOrAfter && beforeEnd {
				iter.key = key
				value, err := node.Value()
				if err != nil {
					iter.fail(err)
					return
				}
				iter.value = value
				return
			}
		} else {
			// push children to stack
			if iter.ascending {
				if beforeEnd {
					iter.pushStack(node.Right())
				}
				if afterStart {
					iter.pushStack(node.Left())
				}
			} else {
				if afterStart {
					iter.pushStack(node.Left())
				}
				if beforeEnd {
					iter.pushStack(node.Right())
				}
			}
		}
	}

	iter.valid = false
}

func (iter *Iterator) pushStack(nodePtr *NodePointer) {
	node, err := nodePtr.Resolve()
	if err != nil {
		iter.fail(err)
		return
	}
	iter.stack = append(iter.stack, node)
}

func (iter *Iterator) fail(err error) {
	iter.valid = false
	iter.err = err
}

// Close implements dbm.Iterator
func (iter *Iterator) Close() error {
	iter.valid = false
	iter.stack = nil
	return nil
}
