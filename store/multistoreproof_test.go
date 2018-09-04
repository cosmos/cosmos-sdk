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

	var storeInfos []storeInfo

	gocRootHash, _ := hex.DecodeString("62c171bb022e47d1f745608ff749e676dbd25f78")
	storeInfos = append(storeInfos, storeInfo{
		Name: "gov",
		Core: storeCore{
			CommitID: CommitID{
				Version: 689,
				Hash:    gocRootHash,
			},
		},
	})

	storeInfos = append(storeInfos, storeInfo{
		Name: "main",
		Core: storeCore{
			CommitID: CommitID{
				Version: 689,
				Hash:    nil,
			},
		},
	})

	accRootHash, _ := hex.DecodeString("ea5d468431015c2cd6295e9a0bb1fc0e49033828")
	storeInfos = append(storeInfos, storeInfo{
		Name: "acc",
		Core: storeCore{
			CommitID: CommitID{
				Version: 689,
				Hash:    accRootHash,
			},
		},
	})

	storeInfos = append(storeInfos, storeInfo{
		Name: "ibc",
		Core: storeCore{
			CommitID: CommitID{
				Version: 689,
				Hash:    nil,
			},
		},
	})

	stakeRootHash, _ := hex.DecodeString("987d1d27b8771d93aa3691262f661d2c85af7ca4")
	storeInfos = append(storeInfos, storeInfo{
		Name: "stake",
		Core: storeCore{
			CommitID: CommitID{
				Version: 689,
				Hash:    stakeRootHash,
			},
		},
	})

	slashingRootHash, _ := hex.DecodeString("388ee6e5b11f367069beb1eefd553491afe9d73e")
	storeInfos = append(storeInfos, storeInfo{
		Name: "slashing",
		Core: storeCore{
			CommitID: CommitID{
				Version: 689,
				Hash:    slashingRootHash,
			},
		},
	})

	commitHash, err := VerifyMultiStoreCommitInfo(storeName, storeInfos, appHash)
	assert.Nil(t, err)
	assert.Equal(t, commitHash, substoreRootHash)

	appHash, _ = hex.DecodeString("29de216bf5e2531c688de36caaf024cd3bb09ee3")

	_, err = VerifyMultiStoreCommitInfo(storeName, storeInfos, appHash)
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
