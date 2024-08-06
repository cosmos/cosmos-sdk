package proof

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/store/v2/internal/encoding"
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
		Name      []byte
		CommitID  CommitID
		Structure string
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

	rootHash, _, _ := ci.GetStoreProof([]byte{})
	return rootHash
}

// GetStoreCommitID returns the CommitID for the given store key.
func (ci *CommitInfo) GetStoreCommitID(storeKey []byte) CommitID {
	for _, si := range ci.StoreInfos {
		if bytes.Equal(si.Name, storeKey) {
			return si.CommitID
		}
	}
	return CommitID{}
}

// GetStoreProof takes in a storeKey and returns a proof of the store key in addition
// to the root hash it should be proved against. If an empty string is provided, the first
// store based on lexographical ordering will be proved.
func (ci *CommitInfo) GetStoreProof(storeKey []byte) ([]byte, *CommitmentOp, error) {
	sort.Slice(ci.StoreInfos, func(i, j int) bool {
		return bytes.Compare(ci.StoreInfos[i].Name, ci.StoreInfos[j].Name) < 0
	})

	isEmpty := len(storeKey) == 0
	index := -1
	leaves := make([][]byte, len(ci.StoreInfos))
	for i, si := range ci.StoreInfos {
		var err error
		leaves[i], err = LeafHash(si.Name, si.GetHash())
		if err != nil {
			return nil, nil, err
		}
		if !isEmpty && bytes.Equal(si.Name, storeKey) {
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

// encodedSize returns the encoded size of CommitInfo for preallocation in Marshal.
func (ci *CommitInfo) encodedSize() int {
	size := encoding.EncodeUvarintSize(ci.Version)
	size += encoding.EncodeVarintSize(ci.Timestamp.UnixNano())
	size += encoding.EncodeUvarintSize(uint64(len(ci.StoreInfos)))
	for _, storeInfo := range ci.StoreInfos {
		size += encoding.EncodeBytesSize(storeInfo.Name)
		size += encoding.EncodeBytesSize(storeInfo.CommitID.Hash)
		size += encoding.EncodeBytesSize([]byte(storeInfo.Structure))
	}
	return size
}

// Marshal returns the encoded byte representation of CommitInfo.
// NOTE: CommitInfo is encoded as follows:
// - version (uvarint)
// - timestamp (varint)
// - number of stores (uvarint)
// - for each store:
//   - store name (bytes)
//   - store hash (bytes)
//   - store commit structure (bytes)
func (ci *CommitInfo) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(ci.encodedSize())

	if err := encoding.EncodeUvarint(&buf, ci.Version); err != nil {
		return nil, err
	}
	if err := encoding.EncodeVarint(&buf, ci.Timestamp.UnixNano()); err != nil {
		return nil, err
	}
	if err := encoding.EncodeUvarint(&buf, uint64(len(ci.StoreInfos))); err != nil {
		return nil, err
	}
	for _, si := range ci.StoreInfos {
		if err := encoding.EncodeBytes(&buf, si.Name); err != nil {
			return nil, err
		}
		if err := encoding.EncodeBytes(&buf, si.CommitID.Hash); err != nil {
			return nil, err
		}
		if err := encoding.EncodeBytes(&buf, []byte(si.Structure)); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Unmarshal unmarshals the encoded byte representation of CommitInfo.
func (ci *CommitInfo) Unmarshal(buf []byte) error {
	// Version
	version, n, err := encoding.DecodeUvarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]
	ci.Version = version
	// Timestamp
	timestamp, n, err := encoding.DecodeVarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]
	ci.Timestamp = time.Unix(timestamp/int64(time.Second), timestamp%int64(time.Second))
	// StoreInfos
	storeInfosLen, n, err := encoding.DecodeUvarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]
	ci.StoreInfos = make([]StoreInfo, storeInfosLen)
	for i := 0; i < int(storeInfosLen); i++ {
		// Name
		name, n, err := encoding.DecodeBytes(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
		ci.StoreInfos[i].Name = name
		// CommitID
		hash, n, err := encoding.DecodeBytes(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
		// Structure
		structure, n, err := encoding.DecodeBytes(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
		ci.StoreInfos[i].Structure = string(structure)

		ci.StoreInfos[i].CommitID = CommitID{
			Hash:    hash,
			Version: ci.Version,
		}
	}

	return nil
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
