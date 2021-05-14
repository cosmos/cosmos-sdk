package smt_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smtstore "github.com/cosmos/cosmos-sdk/store/smt"
	"github.com/lazyledger/smt"
	dbm "github.com/tendermint/tm-db"
)

func TestProofOpInterface(t *testing.T) {
	hasher := sha256.New()
	tree := smt.NewSparseMerkleTree(dbm.NewMemDB(), hasher)
	key := []byte("foo")
	value := []byte("bar")
	root, err := tree.Update(key, value)
	require.NoError(t, err)
	require.NotEmpty(t, root)

	proof, err := tree.Prove(key)
	require.True(t, smt.VerifyProof(proof, root, key, value, hasher))

	storeProofOp := smtstore.NewProofOp(root, key, smtstore.SHA256, proof)
	require.NotNil(t, storeProofOp)
	// inclusion proof
	r, err := storeProofOp.Run([][]byte{value})
	assert.NoError(t, err)
	assert.NotEmpty(t, r)
	assert.Equal(t, root, r[0])

	// inclusion proof - wrong value - should fail
	r, err = storeProofOp.Run([][]byte{key})
	assert.Error(t, err)
	assert.Empty(t, r)

	// exclusion proof - should fail
	r, err = storeProofOp.Run([][]byte{})
	assert.Error(t, err)
	assert.Empty(t, r)

	// invalid request - should fail
	r, err = storeProofOp.Run([][]byte{key, key})
	assert.Error(t, err)
	assert.Empty(t, r)

	// encode
	tmProofOp := storeProofOp.ProofOp()
	assert.NotNil(t, tmProofOp)
	assert.Equal(t, smtstore.ProofType, tmProofOp.Type)
	assert.Equal(t, key, tmProofOp.Key, key)
	assert.NotEmpty(t, tmProofOp.Data)

	//decode
	decoded, err := smtstore.ProofDecoder(tmProofOp)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.Equal(t, key, decoded.GetKey())

	// run proof after decoding
	r, err = decoded.Run([][]byte{value})
	assert.NoError(t, err)
	assert.NotEmpty(t, r)
	assert.Equal(t, root, r[0])
}
