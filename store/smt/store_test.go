package smt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/store/smt"
	dbm "github.com/tendermint/tm-db"
)

func TestVersioning(t *testing.T) {
	s := smt.NewStore(dbm.NewMemDB())
	expectedVersion := int64(0)

	s.Set([]byte("foo"), []byte("bar"))
	cid1 := s.Commit()
	expectedVersion++

	assert.Equal(t, expectedVersion, cid1.Version)
	assert.NotEmpty(t, cid1.Hash)

	s.Set([]byte("foobar"), []byte("baz"))
	cid2 := s.Commit()
	expectedVersion++

	assert.Equal(t, expectedVersion, cid2.Version)
	assert.NotEmpty(t, cid2.Hash)
	assert.NotEqual(t, cid1.Hash, cid2.Hash)
}

func TestInitialVersion(t *testing.T) {
	s := smt.NewStore(dbm.NewMemDB())
	expectedVersion := int64(42)

	s.SetInitialVersion(expectedVersion)

	s.Set([]byte("foo"), []byte("foobar"))
	cid := s.Commit()

	assert.Equal(t, expectedVersion, cid.Version)
	assert.NotEmpty(t, cid.Hash)
}
