package branch

import (
	"slices"

	"cosmossdk.io/store/v2"
)

var _ store.Iterator = (*iterator)(nil)

// iterator walks over both the KVStore's changeset, i.e. dirty writes, and the
// parent iterator, which can either be another KVStore or the SS backend, at the
// same time.
//
// Note, writes that happen on the KVStore over an iterator will not affect the
// iterator. This is because when an iterator is created, it takes a current
// snapshot of the changeset.
type iterator struct {
	parentItr store.Iterator
	start     []byte
	end       []byte
	key       []byte
	value     []byte
	keys      []string
	values    store.KVPairs
	reverse   bool
	exhausted bool // exhausted reflects if the parent iterator is exhausted or not
}

// Domain returns the domain of the iterator. The caller must not modify the
// return values.
func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr *iterator) Key() []byte {
	return slices.Clone(itr.key)
}

func (itr *iterator) Value() []byte {
	return slices.Clone(itr.value)
}

func (itr *iterator) Close() {
	itr.key = nil
	itr.value = nil
	itr.keys = nil
	itr.values = nil
	itr.parentItr.Close()
}

func (itr *iterator) Next() bool {
	for {
		switch {
		case itr.exhausted && len(itr.keys) == 0: // exhausted both
			itr.key = nil
			itr.value = nil
			return false

		case itr.exhausted: // exhausted parent iterator but not store (dirty writes) iterator
			nextKey := itr.keys[0]
			nextValue := itr.values[0]

			// pop off the key
			itr.keys[0] = ""
			itr.keys = itr.keys[1:]

			// pop off the value
			itr.values[0].Value = nil
			itr.values = itr.values[1:]

			if nextValue.Value != nil {
				itr.key = []byte(nextKey)
				itr.value = nextValue.Value
				return true
			}

		case len(itr.keys) == 0: // exhausted store (dirty writes) iterator but not parent iterator
			itr.key = itr.parentItr.Key()
			itr.value = itr.parentItr.Value()
			itr.exhausted = !itr.parentItr.Next()

			return true

		default: // parent iterator is not exhausted and we have store (dirty writes) remaining
			dirtyKey := itr.keys[0]
			dirtyVal := itr.values[0]

			parentKey := itr.parentItr.Key()
			parentKeyStr := string(parentKey)

			switch {
			case (!itr.reverse && dirtyKey < parentKeyStr) || (itr.reverse && dirtyKey > parentKeyStr): // dirty key should come before parent's key
				// pop off key
				itr.keys[0] = ""
				itr.keys = itr.keys[1:]

				// pop off value
				itr.values[0].Value = nil
				itr.values = itr.values[1:]

				if dirtyVal.Value != nil {
					itr.key = []byte(dirtyKey)
					itr.value = dirtyVal.Value
					return true
				}

			case (!itr.reverse && parentKeyStr < dirtyKey) || (itr.reverse && parentKeyStr > dirtyKey): // parent's key should come before dirty key
				itr.key = parentKey
				itr.value = itr.parentItr.Value()
				itr.exhausted = !itr.parentItr.Next()
				return true

			default:
				// pop off key
				itr.keys[0] = ""
				itr.keys = itr.keys[1:]

				// pop off value
				itr.values[0].Value = nil
				itr.values = itr.values[1:]

				itr.exhausted = !itr.parentItr.Next()

				if dirtyVal.Value != nil {
					itr.key = []byte(dirtyKey)
					itr.value = dirtyVal.Value
					return true
				}
			}
		}
	}
}

func (itr *iterator) Valid() bool {
	return itr.key != nil && itr.value != nil
}

func (itr *iterator) Error() error {
	return itr.parentItr.Error()
}
