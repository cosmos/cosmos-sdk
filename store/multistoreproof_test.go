package store

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/iavl"
	cmn "github.com/tendermint/tendermint/libs/common"
	"testing"
)

func TestVerifyMultiStoreCommitInfo(t *testing.T) {
	appHash, _ := hex.DecodeString("ebf3c1fb724d3458023c8fefef7b33add2fc1e84")

	substoreRootHash, _ := hex.DecodeString("ea5d468431015c2cd6295e9a0bb1fc0e49033828")
	storeName := "acc"

	var multiStoreCommitInfo []SubstoreCommitID

	gocRootHash, _ := hex.DecodeString("62c171bb022e47d1f745608ff749e676dbd25f78")
	multiStoreCommitInfo = append(multiStoreCommitInfo, SubstoreCommitID{
		Name:       "gov",
		Version:    689,
		CommitHash: gocRootHash,
	})

	multiStoreCommitInfo = append(multiStoreCommitInfo, SubstoreCommitID{
		Name:       "main",
		Version:    689,
		CommitHash: nil,
	})

	accRootHash, _ := hex.DecodeString("ea5d468431015c2cd6295e9a0bb1fc0e49033828")
	multiStoreCommitInfo = append(multiStoreCommitInfo, SubstoreCommitID{
		Name:       "acc",
		Version:    689,
		CommitHash: accRootHash,
	})

	multiStoreCommitInfo = append(multiStoreCommitInfo, SubstoreCommitID{
		Name:       "ibc",
		Version:    689,
		CommitHash: nil,
	})

	stakeRootHash, _ := hex.DecodeString("987d1d27b8771d93aa3691262f661d2c85af7ca4")
	multiStoreCommitInfo = append(multiStoreCommitInfo, SubstoreCommitID{
		Name:       "stake",
		Version:    689,
		CommitHash: stakeRootHash,
	})

	slashingRootHash, _ := hex.DecodeString("388ee6e5b11f367069beb1eefd553491afe9d73e")
	multiStoreCommitInfo = append(multiStoreCommitInfo, SubstoreCommitID{
		Name:       "slashing",
		Version:    689,
		CommitHash: slashingRootHash,
	})

	commitHash, err := VerifyMultiStoreCommitInfo(storeName, multiStoreCommitInfo, appHash)
	assert.Nil(t, err)
	assert.Equal(t, commitHash, substoreRootHash)

	appHash, _ = hex.DecodeString("29de216bf5e2531c688de36caaf024cd3bb09ee3")

	_, err = VerifyMultiStoreCommitInfo(storeName, multiStoreCommitInfo, appHash)
	assert.Error(t, err, "appHash doesn't match to the merkle root of multiStoreCommitInfo")
}

func TestVerifyRangeProof(t *testing.T) {
	tree := iavl.NewTree(nil, 0)

	rand := cmn.NewRand()
	rand.Seed(0) // for determinism
	for _, ikey := range []byte{0x11, 0x32, 0x50, 0x72, 0x99} {
		key := []byte{ikey}
		tree.Set(key, []byte(rand.Str(8)))
	}

	root := tree.Hash()

	key := []byte{0x32}
	val, proof, err := tree.GetWithProof(key)
	assert.Nil(t, err)
	assert.NotEmpty(t, val)
	assert.NotEmpty(t, proof)
	err = VerifyRangeProof(key, val, root, proof)
	assert.Nil(t, err)

	key = []byte{0x40}
	val, proof, err = tree.GetWithProof(key)
	assert.Nil(t, err)
	assert.Empty(t, val)
	assert.NotEmpty(t, proof)
	err = VerifyRangeProof(key, val, root, proof)
	assert.Nil(t, err)
}