package proof

import (
	"fmt"
	"sort"
	"strings"
)

func (si StoreInfo) GetHash() []byte {
	return si.CommitId.Hash
}

// Hash returns the root hash of all committed stores represented by CommitInfo,
// sorted by store name/key.
func (ci *CommitInfo) Hash() []byte {
	if len(ci.StoreInfos) == 0 {
		return nil
	}

	if len(ci.CommitHash) != 0 {
		return ci.CommitHash
	}

	rootHash, _, _ := ci.GetStoreProof([]byte{})
	return rootHash
}

// GetStoreCommitID returns the CommitID for the given store key.
func (ci *CommitInfo) GetStoreCommitID(storeKey []byte) *CommitID {
	for _, si := range ci.StoreInfos {
		if strings.EqualFold(si.Name, string(storeKey)) {
			return si.CommitId
		}
	}
	return &CommitID{}
}

// GetStoreProof takes in a storeKey and returns a proof of the store key in addition
// to the root hash it should be proved against. If an empty string is provided, the first
// store based on lexographical ordering will be proved.
func (ci *CommitInfo) GetStoreProof(storeKey []byte) ([]byte, *CommitmentOp, error) {
	sort.Slice(ci.StoreInfos, func(i, j int) bool {
		return strings.Compare(ci.StoreInfos[i].Name, ci.StoreInfos[j].Name) < 0
	})

	isEmpty := len(storeKey) == 0
	index := -1
	leaves := make([][]byte, len(ci.StoreInfos))
	for i, si := range ci.StoreInfos {
		var err error
		leaves[i], err = LeafHash([]byte(si.Name), si.GetHash())
		if err != nil {
			return nil, nil, err
		}
		if !isEmpty && strings.EqualFold(si.Name, string(storeKey)) {
			index = i
		}
	}

	if index == -1 {
		if isEmpty {
			index = 0
		} else {
			return nil, nil, fmt.Errorf("store key %s not found", storeKey)
		}
	}

	rootHash, inners := ProofFromByteSlices(leaves, index)
	commitmentOp := ConvertCommitmentOp(inners, storeKey, ci.StoreInfos[index].GetHash())
	return rootHash, &commitmentOp, nil
}

func (ci *CommitInfo) CommitID() *CommitID {
	return &CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}

func (cid *CommitID) String() string {
	return fmt.Sprintf("CommitID{%v:%X}", cid.Hash, cid.Version)
}

func (cid *CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}
