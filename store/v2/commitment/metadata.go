package commitment

import (
	"bytes"
	"errors"
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

// MetadataStore is a store for metadata related to the commitment store.
type MetadataStore struct {
	kv corestore.KVStoreWithBatch
}

// NewMetadataStore creates a new MetadataStore.
func NewMetadataStore(kv corestore.KVStoreWithBatch) *MetadataStore {
	return &MetadataStore{
		kv: kv,
	}
}

// GetLatestVersion returns the latest committed version.
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

// GetCommitInfo returns the commit info for the given version.
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

	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func (m *MetadataStore) flushRemovedStoreKeys(version uint64, storeKeys []string) (err error) {
	batch := m.kv.NewBatch()
	defer func() {
		cErr := batch.Close()
		if err == nil {
			err = cErr
		}
	}()

	for _, storeKey := range storeKeys {
		key := []byte(fmt.Sprintf("%s%s", encoding.BuildPrefixWithVersion(removedStoreKeyPrefix, version), storeKey))
		if err := batch.Set(key, []byte{}); err != nil {
			return err
		}
	}
	return batch.Write()
}

func (m *MetadataStore) GetRemovedStoreKeys(version uint64) (storeKeys [][]byte, err error) {
	end := encoding.BuildPrefixWithVersion(removedStoreKeyPrefix, version+1)
	iter, err := m.kv.Iterator([]byte(removedStoreKeyPrefix), end)
	if err != nil {
		return nil, err
	}
	defer func() {
		if ierr := iter.Close(); ierr != nil {
			err = ierr
		}
	}()

	for ; iter.Valid(); iter.Next() {
		storeKey := iter.Key()[len(end):]
		storeKeys = append(storeKeys, storeKey)
	}
	return storeKeys, nil
}

func (m *MetadataStore) deleteRemovedStoreKeys(version uint64, removeStore func(storeKey []byte, version uint64) error) (err error) {
	removedStoreKeys, err := m.GetRemovedStoreKeys(version)
	if err != nil {
		return err
	}
	if len(removedStoreKeys) == 0 {
		return nil
	}

	batch := m.kv.NewBatch()
	defer func() {
		err = errors.Join(err, batch.Close())
	}()
	for _, storeKey := range removedStoreKeys {
		if err := removeStore(storeKey, version); err != nil {
			return err
		}
		if err := batch.Delete(storeKey); err != nil {
			return err
		}
	}

	return batch.Write()
}

func (m *MetadataStore) deleteCommitInfo(version uint64) error {
	cInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
	return m.kv.Delete(cInfoKey)
}
