package types

import (
	"fmt"
	"time"

	"cosmossdk.io/store/v2/internal/maps"
)

type (
	// CommitInfo defines commit information used by the multi-store when committing
	// a version/height.
	CommitInfo struct {
		Version    uint64
		StoreInfos []StoreInfo
		Timestamp  time.Time
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
func (ci CommitInfo) Hash() []byte {
	if len(ci.StoreInfos) == 0 {
		return nil
	}

	rootHash, _, _ := maps.ProofsFromMap(ci.toMap())
	return rootHash
}

func (ci CommitInfo) toMap() map[string][]byte {
	m := make(map[string][]byte, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo.GetHash()
	}

	return m
}

func (ci CommitInfo) CommitID() CommitID {
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
