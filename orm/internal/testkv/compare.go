package testkv

import (
	"bytes"

	"gotest.tools/v3/assert"

	"cosmossdk.io/core/store"
	"cosmossdk.io/orm/model/ormtable"
)

func AssertBackendsEqual(t assert.TestingT, b1, b2 ormtable.Backend) {
	it1, err := b1.CommitmentStoreReader().Iterator(nil, nil)
	assert.NilError(t, err)

	it2, err := b2.CommitmentStoreReader().Iterator(nil, nil)
	assert.NilError(t, err)

	AssertIteratorsEqual(t, it1, it2)

	it1, err = b1.IndexStoreReader().Iterator(nil, nil)
	assert.NilError(t, err)

	it2, err = b2.IndexStoreReader().Iterator(nil, nil)
	assert.NilError(t, err)

	AssertIteratorsEqual(t, it1, it2)
}

func AssertIteratorsEqual(t assert.TestingT, it1, it2 store.Iterator) {
	for it1.Valid() {
		assert.Assert(t, it2.Valid())
		assert.Assert(t, bytes.Equal(it1.Key(), it2.Key()))
		assert.Assert(t, bytes.Equal(it1.Value(), it2.Value()))
		it1.Next()
		it2.Next()
	}
}
