package commitment

import (
	"bytes"
	"fmt"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/internal/encoding"
	"cosmossdk.io/store/v2/proof"
)

const (
	commitInfoKeyFmt      = "c/%d" // c/<version>
	latestVersionKey      = "c/latest"
	removedStoreKeyPrefix = "c/removed/" // c/removed/<version>/<store-name>
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

func (m *MetadataStore) setLatestVersion(version uint64) error {
	var buf bytes.Buffer
	buf.Grow(encoding.EncodeUvarintSize(version))
	if err := encoding.EncodeUvarint(&buf, version); err != nil {
		return err
	}
	return m.kv.Set([]byte(latestVersionKey), buf.Bytes())
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

func (m *MetadataStore) flushCommitInfo(version uint64, cInfo *proof.CommitInfo) error {
	// do nothing if commit info is nil, as will be the case for an empty, initializing store
	if cInfo == nil {
		return nil
	}

	batch := m.kv.NewBatch()
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
	return batch.Close()
}

func (m *MetadataStore) flushRemovedStoreKeys(version uint64, storeKeys []string) (err error) {
	batch := m.kv.NewBatch()
	defer func() {
		err = batch.Close()
	}()

	for _, storeKey := range storeKeys {
		key := []byte(fmt.Sprintf("%s%s", encoding.BuildPrefixWithVersion(removedStoreKeyPrefix, version), storeKey))
		if err := batch.Set(key, []byte{}); err != nil {
			return err
		}
	}
	return batch.WriteSync()
}

func (m *MetadataStore) deleteRemovedStoreKeys(version uint64, fn func(storeKey []byte) error) (err error) {
	batch := m.kv.NewBatch()
	defer func() {
		if berr := batch.Close(); berr != nil {
			err = berr
		}
	}()

	end := encoding.BuildPrefixWithVersion(removedStoreKeyPrefix, version+1)
	iter, err := m.kv.Iterator([]byte(removedStoreKeyPrefix), end)
	if err != nil {
		return err
	}
	defer func() {
		if ierr := iter.Close(); ierr != nil {
			err = ierr
		}
	}()

	for ; iter.Valid(); iter.Next() {
		storeKey := iter.Key()[len(end):]
		if err := fn(storeKey); err != nil {
			return err
		}
		if err := batch.Delete(iter.Key()); err != nil {
			return nil
		}
	}

	return batch.WriteSync()
}

func (m *MetadataStore) deleteCommitInfo(version uint64) error {
	cInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
	return m.kv.Delete(cInfoKey)
}
