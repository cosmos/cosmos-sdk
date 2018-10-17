package basemulti

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto/merkle"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/types"
)

type Store struct {
	db dbm.DB

	kvstores     map[types.KVStoreKey]types.CommitKVStore
	kvkeysByName map[string]types.KVStoreKey

	pruning types.PruningStrategy
}

func (store *Store) MountKVStoreWithDB(key types.KVStoreKey, db dbm.DB) {
	if key == nil {
		panic("MountStoreWithDB() key cannot be nil")
	}
	if _, ok := store.kvkeysByName[key.Name()]; ok {
		panic(fmt.Sprintf("Store duplicate store key %v", key))
	}

	store.kvkeysByName[key.Name()] = key
}

func (store *Store) GetCommitKVStore(key types.KVStoreKey) types.CommitKVStore {
	return store.kvstores[key]
}

func (store *Store) LoadMultiStoreVersion(ver int64) (err error) {
	// Convert StoreInfos slice to map
	var lastCommitID types.CommitID
	infos := make(map[types.KVStoreKey]storeInfo)
	if ver != 0 {
		// Get commitInfo
		cInfo, err := getCommitInfo(store.db, ver)
		if err != nil {
			return err
		}

		for _, sInfo := range cInfo.storeInfos {
			infos[store.nameToKVKey(sInfo.Name)] = sInfo
		}

		lastCommitID = cInfo.CommitID()
	}

	for _, key := range store.kvkeysByName {
		var id types.CommitID
		if info, ok := infos[key]; ok {
			id = info.Core.CommitID
		}
		kvstore := key.NewStore()
		db := dbm.NewPrefixDB(store.db, []byte("s/k:"+key.Name()+"/"))
		err = kvstore.LoadKVStoreVersion(db, id)
		if err != nil {
			return
		}

		kvstore.SetPruning(store.pruning)
	}
}

func (store *Store) nameToKVKey(name string) types.KVStoreKey {
	for key := range kvstores {
		if key.Name() == name {
			return key
		}
	}
}

// -------------------------------
// storeInfo

// storeInfo contains the name and core reference for an
// underlying store. It is the leaf of the Stores top
// level simple merkle tree

type storeInfo struct {
	Name string
	Core storeCore
}

type storeCore struct {
	CommitID types.CommitID
	// ... maybe add more state
}

// ------------------------------
// commitInfo

// NOTE: keep commitInfo a simple immutable struct.
type commitInfo struct {
	// Version
	Version int64

	// types.Store info for
	storeInfos []storeInfo
}

// Hash returns the simple merkle root hash of the stores sorted by name.
func (ci commitInfo) Hash() []byte {
	// TODO cache to ci.hash []byte
	m := make(map[string]merkle.Hasher, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo
	}
	return merkle.SimpleHashFromMap(m)
}

func (ci commitInfo) CommitID() types.CommitID {
	return types.CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}

// -------------------------------
// Misc.

func getLatestVestoreion(db dbm.DB) (latest int64) {
	latestBytes := db.Get([]byte("s/latest"))
	if latestBytes == nil {
		return 0
	}
	cdc.MustUnmarshalBinary(latestBytes, &latest)
	return
}

func getCommitInfo(db dbm.DB, ver int64) (cInfo commitInfo, err error) {
	cInfoBytes := db.Get([]byte(fmt.Sprintf("s/%d", ver)))
	if cInfoBytes == nil {
		err = fmt.Errorf("failed to get Store: no data")
	}

	err = cdc.UnmarshalBinary(cInfoBytes, &cInfo)
	if err != nil {
		err = fmt.Errorf("failed to get Store: %v", err)
	}
	return
}
