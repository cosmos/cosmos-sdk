package smt_test

import (
	"crypto/sha256"
	"testing"

	ics23 "github.com/confio/ics23/go"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	store "github.com/cosmos/cosmos-sdk/store/v2/smt"
)

func TestProofICS23(t *testing.T) {
	txn := memdb.NewDB().ReadWriter()
	s := store.NewStore(txn)
	// pick keys whose hashes begin with different bits
	key00 := []byte("foo")  // 00101100 = sha256(foo)[0]
	key01 := []byte("bill") // 01100010
	key10 := []byte("baz")  // 10111010
	key11 := []byte("bar")  // 11111100
	path00 := sha256.Sum256(key00)
	path01 := sha256.Sum256(key01)
	path10 := sha256.Sum256(key10)
	val1 := []byte("0")
	val2 := []byte("1")

	s.Set(key01, val1)

	// Membership
	proof, err := s.GetProofICS23(key01)
	assert.NoError(t, err)
	nonexist := proof.GetNonexist()
	assert.Nil(t, nonexist)
	exist := proof.GetExist()
	assert.NotNil(t, exist)
	assert.Equal(t, 0, len(exist.Path))
	assert.NoError(t, exist.Verify(ics23.SmtSpec, s.Root(), path01[:], val1))

	// Non-membership
	proof, err = s.GetProofICS23(key00) // When leaf is leftmost node
	assert.NoError(t, err)
	nonexist = proof.GetNonexist()
	assert.NotNil(t, nonexist)
	assert.Nil(t, nonexist.Left)
	assert.Equal(t, path00[:], nonexist.Key)
	assert.NotNil(t, nonexist.Right)
	assert.Equal(t, 0, len(nonexist.Right.Path))
	assert.NoError(t, nonexist.Verify(ics23.SmtSpec, s.Root(), path00[:]))

	proof, err = s.GetProofICS23(key10) // When rightmost
	assert.NoError(t, err)
	nonexist = proof.GetNonexist()
	assert.NotNil(t, nonexist)
	assert.NotNil(t, nonexist.Left)
	assert.Equal(t, 0, len(nonexist.Left.Path))
	assert.Nil(t, nonexist.Right)
	assert.NoError(t, nonexist.Verify(ics23.SmtSpec, s.Root(), path10[:]))
	badNonexist := nonexist

	s.Set(key11, val2)

	proof, err = s.GetProofICS23(key10) // In between two keys
	assert.NoError(t, err)
	nonexist = proof.GetNonexist()
	assert.NotNil(t, nonexist)
	assert.Equal(t, path10[:], nonexist.Key)
	assert.NotNil(t, nonexist.Left)
	assert.Equal(t, 1, len(nonexist.Left.Path))
	assert.NotNil(t, nonexist.Right)
	assert.Equal(t, 1, len(nonexist.Right.Path))
	assert.NoError(t, nonexist.Verify(ics23.SmtSpec, s.Root(), path10[:]))

	// Make sure proofs work with a loaded store
	root := s.Root()
	s = store.LoadStore(txn, root)
	proof, err = s.GetProofICS23(key10)
	assert.NoError(t, err)
	nonexist = proof.GetNonexist()
	assert.Equal(t, path10[:], nonexist.Key)
	assert.NotNil(t, nonexist.Left)
	assert.Equal(t, 1, len(nonexist.Left.Path))
	assert.NotNil(t, nonexist.Right)
	assert.Equal(t, 1, len(nonexist.Right.Path))
	assert.NoError(t, nonexist.Verify(ics23.SmtSpec, s.Root(), path10[:]))

	// Invalid proofs should fail to verify
	badExist := exist // expired proof
	assert.Error(t, badExist.Verify(ics23.SmtSpec, s.Root(), path01[:], val1))

	badExist = nonexist.Left
	badExist.Key = key01 // .Key must contain key path
	assert.Error(t, badExist.Verify(ics23.SmtSpec, s.Root(), path01[:], val1))

	badExist = nonexist.Left
	badExist.Path[0].Prefix = []byte{0} // wrong inner node prefix
	assert.Error(t, badExist.Verify(ics23.SmtSpec, s.Root(), path01[:], val1))

	badExist = nonexist.Left
	badExist.Path = []*ics23.InnerOp{} // empty path
	assert.Error(t, badExist.Verify(ics23.SmtSpec, s.Root(), path01[:], val1))

	assert.Error(t, badNonexist.Verify(ics23.SmtSpec, s.Root(), path10[:]))

	badNonexist = nonexist
	badNonexist.Key = key10
	assert.Error(t, badNonexist.Verify(ics23.SmtSpec, s.Root(), path10[:]))
}
