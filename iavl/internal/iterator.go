package internal

type Iterator struct {
	// domain of iteration, end is exclusive
	start, end []byte
	ascending  bool

	// cache the next key-value pair
	key, value []byte

	err   error
	valid bool

	stack []*NodePointer
}

func NewIterator(start, end []byte, ascending bool, root *NodePointer) *Iterator {
	iter := &Iterator{
		start:     start,
		end:       end,
		ascending: ascending,
		valid:     true,
	}

	if root != nil {
		iter.stack = []*NodePointer{root}
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
	return iter.err
}

// Key implements dbm.Iterator
func (iter *Iterator) Key() []byte {
	return iter.key
}

// Value implements dbm.Iterator
func (iter *Iterator) Value() []byte {
	return iter.value
}

// Next implements dbm.Iterator
func (iter *Iterator) Next() {
	// TODO we can do optimizations with CmpKey to avoid loading full keys when possible
	// TODO we can keep a stack of Node's and Pin's to avoid SafeCopy when not requested by the user
	if !iter.valid {
		return
	}

	for len(iter.stack) > 0 {
		// pop node
		nodePtr := iter.stack[len(iter.stack)-1]
		iter.stack = iter.stack[:len(iter.stack)-1]

		node, pin, err := nodePtr.Resolve()
		// TODO this defer is done in a for loop which isn't really correct, but it is also somewhat harmless, we could fix in the future
		defer pin.Unpin()
		if err != nil {
			iter.fail(err)
			return
		}

		startCmp, err := node.CmpKey(iter.start)
		if err != nil {
			iter.fail(err)
			return
		}
		afterStart := iter.start == nil || startCmp > 0
		endCmp, err := node.CmpKey(iter.end)
		if err != nil {
			iter.fail(err)
			return
		}
		beforeEnd := iter.end == nil || endCmp < 0

		if node.IsLeaf() {
			startOrAfter := afterStart || startCmp == 0
			if startOrAfter && beforeEnd {
				key, err := node.Key()
				if err != nil {
					iter.fail(err)
					return
				}
				iter.key = key.SafeCopy()
				value, err := node.Value()
				if err != nil {
					iter.fail(err)
					return
				}
				iter.value = value.SafeCopy()
				return
			}
		} else {
			// push children to stack
			if iter.ascending {
				if beforeEnd {
					iter.stack = append(iter.stack, node.Right())
				}
				if afterStart {
					iter.stack = append(iter.stack, node.Left())
				}
			} else {
				if afterStart {
					iter.stack = append(iter.stack, node.Left())
				}
				if beforeEnd {
					iter.stack = append(iter.stack, node.Right())
				}
			}
		}
	}

	iter.valid = false
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
