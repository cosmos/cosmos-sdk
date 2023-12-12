package store

import (
	"bytes"
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

func (ci *CommitInfo) encodedSize() int {
	size := EncodeUvarintSize(ci.Version)
	size += EncodeVarintSize(ci.Timestamp.UnixNano())
	size += EncodeUvarintSize(uint64(len(ci.StoreInfos)))
	for _, storeInfo := range ci.StoreInfos {
		size += EncodeBytesSize([]byte(storeInfo.Name))
		size += EncodeBytesSize(storeInfo.CommitID.Hash)
	}
	return size
}

func (ci *CommitInfo) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(ci.encodedSize())

	if err := EncodeUvarint(&buf, ci.Version); err != nil {
		return nil, err
	}
	if err := EncodeVarint(&buf, ci.Timestamp.UnixNano()); err != nil {
		return nil, err
	}
	if err := EncodeUvarint(&buf, uint64(len(ci.StoreInfos))); err != nil {
		return nil, err
	}
	for _, si := range ci.StoreInfos {
		if err := EncodeBytes(&buf, []byte(si.Name)); err != nil {
			return nil, err
		}
		if err := EncodeBytes(&buf, si.CommitID.Hash); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (ci *CommitInfo) Unmarshal(buf []byte) error {
	// Version
	version, n, err := DecodeUvarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]
	ci.Version = version
	// Timestamp
	timestamp, n, err := DecodeVarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]
	ci.Timestamp = time.Unix(timestamp/int64(time.Second), timestamp%int64(time.Second))
	// StoreInfos
	storeInfosLen, n, err := DecodeVarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]
	ci.StoreInfos = make([]StoreInfo, storeInfosLen)
	for i := 0; i < int(storeInfosLen); i++ {
		// Name
		name, n, err := DecodeBytes(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
		ci.StoreInfos[i].Name = string(name)
		// CommitID
		hash, n, err := DecodeBytes(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
		ci.StoreInfos[i].CommitID = CommitID{
			Hash:    hash,
			Version: ci.Version,
		}
	}

	return nil
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

func (cid CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}
