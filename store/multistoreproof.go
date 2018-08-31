package store

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/tendermint/iavl"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// MultiStoreProof defines a collection of store proofs in a multi-store
type MultiStoreProof struct {
	StoreInfos []storeInfo
	StoreName  string
	RangeProof iavl.RangeProof
}

// buildMultiStoreProof build MultiStoreProof based on iavl proof and storeInfos
func buildMultiStoreProof(iavlProof []byte, storeName string, storeInfos []storeInfo) ([]byte, error) {
	var rangeProof iavl.RangeProof
	err := cdc.UnmarshalBinary(iavlProof, &rangeProof)
	if err != nil {
		return nil, err
	}

	msp := MultiStoreProof{
		StoreInfos: storeInfos,
		StoreName:  storeName,
		RangeProof: rangeProof,
	}

	proof, err := cdc.MarshalBinary(msp)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// VerifyMultiStoreCommitInfo verify multiStoreCommitInfo against appHash
func VerifyMultiStoreCommitInfo(storeName string, storeInfos []storeInfo, appHash []byte) ([]byte, error) {
	var substoreCommitHash []byte
	var height int64
	for _, storeInfo := range storeInfos {
		if storeInfo.Name == storeName {
			substoreCommitHash = storeInfo.Core.CommitID.Hash
			height = storeInfo.Core.CommitID.Version
		}
	}
	if len(substoreCommitHash) == 0 {
		return nil, cmn.NewError("failed to get substore root commit hash by store name")
	}

	ci := commitInfo{
		Version:    height,
		StoreInfos: storeInfos,
	}

	if !bytes.Equal(appHash, ci.Hash()) {
		return nil, cmn.NewError("the merkle root of multiStoreCommitInfo doesn't equal to appHash")
	}
	return substoreCommitHash, nil
}

// VerifyRangeProof verify iavl RangeProof
func VerifyRangeProof(key, value []byte, substoreCommitHash []byte, rangeProof *iavl.RangeProof) error {

	// Validate the proof to ensure data integrity.
	err := rangeProof.Verify(substoreCommitHash)
	if err != nil {
		return errors.Wrap(err, "proof root hash doesn't equal to substore commit root hash")
	}

	if len(value) != 0 {
		// Verify existence proof
		err = rangeProof.VerifyItem(key, value)
		if err != nil {
			return errors.Wrap(err, "failed in existence verification")
		}
	} else {
		// Verify absence proof
		err = rangeProof.VerifyAbsence(key)
		if err != nil {
			return errors.Wrap(err, "failed in absence verification")
		}
	}

	return nil
}

// RequireProof return whether proof is require for the subpath
func RequireProof(subpath string) bool {
	// Currently, only when query subpath is "/store" or "/key", will proof be included in response.
	// If there are some changes about proof building in iavlstore.go, we must change code here to keep consistency with iavlstore.go:212
	if subpath == "/store" || subpath == "/key" {
		return true
	}
	return false
}
