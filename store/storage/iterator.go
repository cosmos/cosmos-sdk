package storage

import "cosmossdk.io/store/v2"

var _ store.Iterator = (*iterator)(nil)

type iterator struct {
	store.Iterator

	db          *Database
	key, value  []byte
	keys        []string
	values      []dbEntry
	initialized bool
	exhausted   bool
	valid       bool
	err         error
}

func newIterator(
	db *Database,
	storeKey string,
	version uint64,
	start, end []byte,
	reverse bool,
	keys []string,
	values []dbEntry,
) (*iterator, error) {
	var (
		dbIter store.Iterator
		err    error
	)

	if reverse {
		dbIter, err = db.vdb.NewReverseIterator(storeKey, version, start, end)
	} else {
		dbIter, err = db.vdb.NewIterator(storeKey, version, start, end)
	}
	if err != nil {
		return nil, err
	}

	return &iterator{
		Iterator: dbIter,
		db:       db,
		keys:     keys,
		values:   values,
	}, nil
}

func (itr *iterator) Valid() bool {
	return itr.valid
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted. We must pay careful attention to set the proper values
// based on if the in memory db or the underlying db should be read next
func (itr *iterator) Next() bool {
	// Short-circuit and set an error if the underlying database has been closed.
	if itr.db.isClosed() {
		itr.key = nil
		itr.value = nil
		itr.err = store.ErrClosed
		itr.valid = false
		return itr.Valid()
	}

	if !itr.initialized {
		itr.exhausted = !itr.Iterator.Next()
		itr.initialized = true
	}

	for {
		switch {
		case itr.exhausted && len(itr.keys) == 0:
			itr.key = nil
			itr.value = nil
			itr.valid = false
			return itr.Valid()

		case itr.exhausted:
			nextKey := itr.keys[0]
			nextValue := itr.values[0]

			itr.keys[0] = ""
			itr.keys = itr.keys[1:]
			itr.values[0].value = nil
			itr.values = itr.values[1:]

			if !nextValue.delete {
				itr.key = []byte(nextKey)
				itr.value = nextValue.value
				itr.valid = true
				return itr.Valid()
			}

		case len(itr.keys) == 0:
			itr.key = itr.Iterator.Key()
			itr.value = itr.Iterator.Value()
			itr.exhausted = !itr.Iterator.Next()
			itr.valid = true
			return itr.Valid()

		default:
			memKey := itr.keys[0]
			memValue := itr.values[0]

			dbKey := itr.Iterator.Key()

			dbStringKey := string(dbKey)
			switch {
			case memKey < dbStringKey:
				itr.keys[0] = ""
				itr.keys = itr.keys[1:]
				itr.values[0].value = nil
				itr.values = itr.values[1:]

				if !memValue.delete {
					itr.key = []byte(memKey)
					itr.value = memValue.value
					itr.valid = true
					return itr.Valid()
				}

			case dbStringKey < memKey:
				itr.key = dbKey
				itr.value = itr.Iterator.Value()
				itr.exhausted = !itr.Iterator.Next()
				itr.valid = true
				return itr.Valid()

			default:
				itr.keys[0] = ""
				itr.keys = itr.keys[1:]
				itr.values[0].value = nil
				itr.values = itr.values[1:]
				itr.exhausted = !itr.Iterator.Next()

				if !memValue.delete {
					itr.key = []byte(memKey)
					itr.value = memValue.value
					itr.valid = true
					return itr.Valid()
				}
			}
		}
	}
}

func (itr *iterator) Error() error {
	if itr.err != nil {
		return itr.err
	}

	return itr.Iterator.Error()
}

func (itr *iterator) Key() []byte {
	return itr.key
}

func (itr *iterator) Value() []byte {
	return itr.value
}

func (itr *iterator) Close() {
	itr.key = nil
	itr.value = nil
	itr.keys = nil
	itr.values = nil
	itr.Iterator.Close()
}
