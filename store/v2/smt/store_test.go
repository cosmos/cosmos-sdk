package smt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	store "github.com/cosmos/cosmos-sdk/store/v2/smt"
)

func TestGetSetHasDelete(t *testing.T) {
	db := memdb.NewDB()
	s := store.NewStore(db.ReadWriter())

	s.Set([]byte("foo"), []byte("bar"))
	assert.Equal(t, []byte("bar"), s.Get([]byte("foo")))
	assert.Equal(t, true, s.Has([]byte("foo")))
	s.Delete([]byte("foo"))
	assert.Equal(t, false, s.Has([]byte("foo")))

	assert.Panics(t, func() { s.Get(nil) }, "Get(nil key) should panic")
	assert.Panics(t, func() { s.Get([]byte{}) }, "Get(empty key) should panic")
	assert.Panics(t, func() { s.Has(nil) }, "Has(nil key) should panic")
	assert.Panics(t, func() { s.Has([]byte{}) }, "Has(empty key) should panic")
	assert.Panics(t, func() { s.Set(nil, []byte("value")) }, "Set(nil key) should panic")
	assert.Panics(t, func() { s.Set([]byte{}, []byte("value")) }, "Set(empty key) should panic")
	assert.Panics(t, func() { s.Set([]byte("key"), nil) }, "Set(nil value) should panic")
}

func TestLoadStore(t *testing.T) {
	db := memdb.NewDB()
	txn := db.ReadWriter()
	s := store.NewStore(txn)

	s.Set([]byte{0}, []byte{0})
	s.Set([]byte{1}, []byte{1})
	s.Delete([]byte{1})
	root := s.Root()

	s = store.LoadStore(txn, root)
	assert.Equal(t, []byte{0}, s.Get([]byte{0}))
	assert.False(t, s.Has([]byte{1}))
	s.Set([]byte{2}, []byte{2})
	assert.NotEqual(t, root, s.Root())
}
