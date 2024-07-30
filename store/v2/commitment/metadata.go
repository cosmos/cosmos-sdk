package commitment

import (
	"bytes"
	"fmt"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/internal/encoding"
	"cosmossdk.io/store/v2/proof"
)

const (
	commitInfoKeyFmt = "c/%d" // c/<version>
	latestVersionKey = "c/latest"
)

type MetadataStore struct {
	kv corestore.KVStoreWithBatch
}

func NewMetadataStore(kv corestore.KVStoreWithBatch) *MetadataStore {
	return &MetadataStore{
		kv: kv,
	}
}

func (m *MetadataStore) GetLatestVersion() (uint64, error) {
	value, err := m.kv.Get([]byte(latestVersionKey))
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, nil
	}

	version, _, err := encoding.DecodeUvarint(value)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (m *MetadataStore) GetCommitInfo(version uint64) (*proof.CommitInfo, error) {
	key := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
	value, err := m.kv.Get(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	cInfo := &proof.CommitInfo{}
	if err := cInfo.Unmarshal(value); err != nil {
		return nil, err
	}

	return cInfo, nil
}

func (m *MetadataStore) flushCommitInfo(version uint64, cInfo *proof.CommitInfo) (err error) {
	// do nothing if commit info is nil, as will be the case for an empty, initializing store
	if cInfo == nil {
		return nil
	}

	batch := m.kv.NewBatch()
	defer func() {
		cErr := batch.Close()
		if err == nil {
			err = cErr
		}
	}()
	cInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
	value, err := cInfo.Marshal()
	if err != nil {
		return err
	}
	if err := batch.Set(cInfoKey, value); err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.Grow(encoding.EncodeUvarintSize(version))
	if err := encoding.EncodeUvarint(&buf, version); err != nil {
		return err
	}
	if err := batch.Set([]byte(latestVersionKey), buf.Bytes()); err != nil {
		return err
	}

	if err := batch.WriteSync(); err != nil {
		return err
	}
	return nil
}

func (m *MetadataStore) deleteCommitInfo(version uint64) error {
	cInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
	return m.kv.Delete(cInfoKey)
}
