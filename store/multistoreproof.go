package store

import (
	"github.com/tendermint/tmlibs/merkle"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tmlibs/merkle/tmhash"
	"encoding/binary"
	"io"
	"bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func VerifyProofForMultiStore(storeName string, substoreRootHash []byte, multiStoreCommitInfo []iavl.MultiStoreCommitID,appHash []byte) (error){
	found :=  false
	var kvPairs cmn.KVPairs
	for _,multiStoreCommitID := range multiStoreCommitInfo {
		if multiStoreCommitID.Name == storeName && bytes.Equal(substoreRootHash,multiStoreCommitID.CommitHash){
			found = true;
		}
		kHash := merkle.SimpleHashFromBytes([]byte(multiStoreCommitID.Name))

		storeInfo := storeInfo{
			Core:storeCore{
				CommitID:sdk.CommitID{
					Version: multiStoreCommitID.Version,
					Hash: multiStoreCommitID.CommitHash,
				},
			},
		}

		kvPairs = append(kvPairs, cmn.KVPair{
			Key:   kHash,
			Value: storeInfo.Hash(),
		})
	}
	if !found {
		return cmn.NewError("Invalid proof, there is no matched multiStore in multiStoreCommitInfo")
	}
	if kvPairs == nil {
		return cmn.NewError("Error in extracting information from multiStoreCommitInfo")
	}
	//sort the kvPair list
	kvPairs.Sort()


	//Rebuild simple merkle hash tree
	var hashList [][]byte
	for _, kvPair := range kvPairs {
		hashResult := kvPairHash(kvPair.Key,kvPair.Value)
		hashList=append(hashList,hashResult)
	}

	if !bytes.Equal(appHash,simpleHashFromHashes(hashList)){
		return cmn.NewError("AppHash doesn't match")
	}

	return nil
}

func BuildStoreInfoAndReturnHash(storeName string, height int64, rootHash []byte) []byte {
	storeInfo := storeInfo{
		Core:storeCore{
			CommitID:sdk.CommitID{
				Version: height,
				Hash: rootHash,
			},
		},
	}
	return kvPairHash(merkle.SimpleHashFromBytes([]byte(storeName)),storeInfo.Hash());
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
		return kvPairHash(left,right)
	}
}

func kvPairHash(part1, part2 []byte) []byte {
	hasher := tmhash.New()

	err := encodeByteSlice(hasher, part1)
	if err != nil {
		panic(err)
	}
	err = encodeByteSlice(hasher, part2)
	if err != nil {
		panic(err)
	}
	return hasher.Sum(nil)
}

func encodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = encodeUvarint(w, uint64(len(bz)))
	if err != nil {
		return
	}
	_, err = w.Write(bz)
	return
}

func encodeUvarint(w io.Writer, i uint64) (err error) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], i)
	_, err = w.Write(buf[0:n])
	return
}