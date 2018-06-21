package store

import (
	"github.com/tendermint/tmlibs/merkle"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tmlibs/merkle/tmhash"
	"encoding/binary"
	"io"
	"bytes"
)

func BuildProofForMultiStore(commitInfo commitInfo, storeName string) (int64, []iavl.SimpleMerkleHashNode, error){
	var kvPairs cmn.KVPairs
	var height int64
	storeNameHash := merkle.SimpleHashFromBytes([]byte(storeName))
	for _, storeInfo := range commitInfo.StoreInfos {
		kHash := merkle.SimpleHashFromBytes([]byte(storeInfo.Name))
		kvPairs = append(kvPairs, cmn.KVPair{
			Key:   kHash,
			Value: storeInfo.Hash(),
		})
		if bytes.Equal(storeNameHash, kHash) {
			height = storeInfo.Core.CommitID.Version
		}
		//hashToStoreInfoMap[storeInfo.Name]=storeInfo
	}
	if kvPairs == nil {
		return 0,nil,cmn.NewError("Error in build kvPairs for commit storeInfos")
	}
	//sort the kvPair list
	kvPairs.Sort()


	//Rebuild simple merkle hash tree
	var hashNodeList []hashNode
	for _, kvPair := range kvPairs {
		hashResult := kvPairHash(kvPair.Key,kvPair.Value)
		hashNodeList=append(hashNodeList,hashNode{
			encounter:	bytes.Equal(storeNameHash, kvPair.Key),
			hash:	hashResult,
		})
	}
	var hashPath hashPath
	if hashNodeList != nil {
		//Find the path from the Merkle root to target store
		simpleHashFromHashes(hashNodeList, &hashPath)
		return height, hashPath.innerHashNodeList, nil
	}
	return 0,nil,cmn.NewError("Failed to get commit proof for multistore")

}

func VerifyProofForMultiStore(appHash,leftHash []byte,storeName string, proof []iavl.SimpleMerkleHashNode) (bool){
	hash := kvPairHash(merkle.SimpleHashFromBytes([]byte(storeName)),leftHash);
	for _,merkleHashNode := range proof {
		if merkleHashNode.IsLeft {
			hash=kvPairHash(hash,merkleHashNode.Hash)
		} else {
			hash=kvPairHash(merkleHashNode.Hash,hash)
		}
	}
	return bytes.Equal(appHash,hash)
}

type hashNode struct {
	encounter   bool
	hash		[]byte
}

type hashPath struct {
	innerHashNodeList		[]iavl.SimpleMerkleHashNode
}

func simpleHashFromHashes(hashes []hashNode, path *hashPath) hashNode {
	// Recursive impl.
	switch len(hashes) {
	case 0:
		return hashNode{}
	case 1:
		return hashes[0]
	default:
		left := simpleHashFromHashes(hashes[:(len(hashes)+1)/2],path)
		right := simpleHashFromHashes(hashes[(len(hashes)+1)/2:],path)
		if left.encounter {
			path.innerHashNodeList = append(path.innerHashNodeList,iavl.SimpleMerkleHashNode{
				IsLeft:true,
				Hash:right.hash,
			})
		} else if right.encounter {

			path.innerHashNodeList = append(path.innerHashNodeList,
				iavl.SimpleMerkleHashNode{
					IsLeft: false,
					Hash:        left.hash,
				})
		}
		return hashNode {
			encounter: 	left.encounter || right.encounter,
			hash:		kvPairHash(left.hash, right.hash),
		}
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