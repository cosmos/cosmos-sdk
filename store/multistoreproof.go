package store

import (
	"bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/crypto/merkle"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// commitID of substores, such as acc store, gov store
type SubstoreCommitID struct {
	Name       string       `json:"name"`
	Version    int64        `json:"version"`
	CommitHash cmn.HexBytes `json:"commit_hash"`
}

// proof of store which have multi substores
type MultiStoreProof struct {
	CommitIDList []SubstoreCommitID `json:"commit_id_list"`
	StoreName    string             `json:"store_name"`
	RangeProof   iavl.RangeProof    `json:"range_proof"`
}

// build MultiStoreProof based on iavl proof and storeInfos
func BuildMultiStoreProof(iavlProof []byte, storeName string, storeInfos []storeInfo) ([]byte, error) {
	var rangeProof iavl.RangeProof
	err := cdc.UnmarshalBinary(iavlProof, &rangeProof)
	if err != nil {
		return nil, err
	}

	var multiStoreProof MultiStoreProof
	for _, storeInfo := range storeInfos {

		commitID := SubstoreCommitID{
			Name:       storeInfo.Name,
			Version:    storeInfo.Core.CommitID.Version,
			CommitHash: storeInfo.Core.CommitID.Hash,
		}
		multiStoreProof.CommitIDList = append(multiStoreProof.CommitIDList, commitID)
	}
	multiStoreProof.StoreName = storeName
	multiStoreProof.RangeProof = rangeProof

	proof, err := cdc.MarshalBinary(multiStoreProof)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// verify multiStoreCommitInfo against appHash
func VerifyMultiStoreCommitInfo(storeName string, multiStoreCommitInfo []SubstoreCommitID, appHash []byte) ([]byte, error) {
	var substoreCommitHash []byte
	var kvPairs cmn.KVPairs
	for _, multiStoreCommitID := range multiStoreCommitInfo {

		if multiStoreCommitID.Name == storeName {
			substoreCommitHash = multiStoreCommitID.CommitHash
		}

		keyHash := []byte(multiStoreCommitID.Name)
		storeInfo := storeInfo{
			Core: storeCore{
				CommitID: sdk.CommitID{
					Version: multiStoreCommitID.Version,
					Hash:    multiStoreCommitID.CommitHash,
				},
			},
		}

		kvPairs = append(kvPairs, cmn.KVPair{
			Key:   keyHash,
			Value: storeInfo.Hash(),
		})
	}
	if len(substoreCommitHash) == 0 {
		return nil, cmn.NewError("failed to get substore root commit hash by store name")
	}
	if kvPairs == nil {
		return nil, cmn.NewError("failed to extract information from multiStoreCommitInfo")
	}
	//sort the kvPair list
	kvPairs.Sort()

	//Rebuild simple merkle hash tree
	var hashList [][]byte
	for _, kvPair := range kvPairs {
		hashResult := merkle.SimpleHashFromTwoHashes(kvPair.Key, kvPair.Value)
		hashList = append(hashList, hashResult)
	}

	if !bytes.Equal(appHash, simpleHashFromHashes(hashList)) {
		return nil, cmn.NewError("The merkle root of multiStoreCommitInfo doesn't equal to appHash")
	}
	return substoreCommitHash, nil
}

// verify iavl proof
func VerifyRangeProof(key, value []byte, substoreCommitHash []byte, rangeProof *iavl.RangeProof) error {

	// Validate the proof to ensure data integrity.
	err := rangeProof.Verify(substoreCommitHash)
	if err != nil {
		return errors.Wrap(err, "proof root hash doesn't equal to substore commit root hash")
	}

	if len(value) != 0 {
		// Validate existence proof
		err = rangeProof.VerifyItem(key, value)
		if err != nil {
			return errors.Wrap(err, "failed in existence verification")
		}
	} else {
		// Validate absence proof
		err = rangeProof.VerifyAbsence(key)
		if err != nil {
			return errors.Wrap(err, "failed in absence verification")
		}
	}

	return nil
}

func simpleHashFromHashes(hashes [][]byte) []byte {
	// Recursive impl.
	switch len(hashes) {
	case 0:
		return nil
	case 1:
		return hashes[0]
	default:
		left := simpleHashFromHashes(hashes[:(len(hashes)+1)/2])
		right := simpleHashFromHashes(hashes[(len(hashes)+1)/2:])
		return merkle.SimpleHashFromTwoHashes(left, right)
	}
}
