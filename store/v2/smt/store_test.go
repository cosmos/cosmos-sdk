package smt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	store "github.com/cosmos/cosmos-sdk/store/v2/smt"
	"github.com/lazyledger/smt"
)

func TestGetSetHasDelete(t *testing.T) {
	s := store.NewStore(smt.NewSimpleMap(), smt.NewSimpleMap())

	s.Set([]byte("foo"), []byte("bar"))
	assert.Equal(t, []byte("bar"), s.Get([]byte("foo")))
	assert.Equal(t, true, s.Has([]byte("foo")))
	s.Delete([]byte("foo"))
	assert.Equal(t, false, s.Has([]byte("foo")))
}
