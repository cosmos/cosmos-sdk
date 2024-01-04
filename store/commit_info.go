package store

import (
	"fmt"
	"sort"
	"time"
)

type (
	// CommitInfo defines commit information used by the multi-store when committing
	// a version/height.
	CommitInfo struct {
		Version    uint64
		StoreInfos []StoreInfo
		Timestamp  time.Time
		CommitHash []byte
	}

	// StoreInfo defines store-specific commit information. It contains a reference
	// between a store name/key and the commit ID.
	StoreInfo struct {
		Name     string
		CommitID CommitID
	}

	// CommitID defines the commitment information when a specific store is
	// committed.
	CommitID struct {
		Version uint64
		Hash    []byte
	}
)

func (si StoreInfo) GetHash() []byte {
	return si.CommitID.Hash
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

	rootHash, _, _ := ci.GetStoreProof("")
	return rootHash
}

func (ci *CommitInfo) GetStoreProof(storeKey string) ([]byte, *CommitmentOp, error) {
	sort.Slice(ci.StoreInfos, func(i, j int) bool {
		return ci.StoreInfos[i].Name < ci.StoreInfos[j].Name
	})

	index := 0
	leaves := make([][]byte, len(ci.StoreInfos))
	for i, si := range ci.StoreInfos {
		var err error
		leaves[i], err = LeafHash([]byte(si.Name), si.GetHash())
		if err != nil {
			return nil, nil, err
		}
		if si.Name == storeKey {
			index = i
		}
	}

	rootHash, inners := ProofFromByteSlices(leaves, index)
	commitmentOp := ConvertCommitmentOp(inners, []byte(storeKey), ci.StoreInfos[index].GetHash())

	return rootHash, &commitmentOp, nil
}

func (ci *CommitInfo) CommitID() CommitID {
	return CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}

func (m *CommitInfo) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (cid CommitID) String() string {
	return fmt.Sprintf("CommitID{%v:%X}", cid.Hash, cid.Version)
}

func (cid CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}
