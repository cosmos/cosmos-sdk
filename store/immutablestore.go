package store

import (
	dbm "github.com/tendermint/tendermint/libs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ImmutableStore struct {
	dbStoreAdapter

	initialized bool
	commitID    CommitID
}

func NewImmutableStore() *ImmutableStore {
	return &ImmutableStore{
		dbStoreAdapter: dbStoreAdapter{dbm.NewMemDB()},
	}
}

func (is *ImmutableStore) Commit() CommitID {
	if !is.initialized {
		panic("Commit() on not initialized ImmutableStore")
	}
	is.commitID.Version += 1
	return is.commitID
}

func (is *ImmutableStore) LastCommitID() CommitID {
	return is.commitID
}

func (is *ImmutableStore) SetPruning(pruning sdk.PruningStrategy) {
	return
}

func (is *ImmutableStore) Set(key, value []byte) {
	panic("Set() called on ImmutableStore")
}

func (is *ImmutableStore) Delete(key []byte) {
	panic("Delete() called on ImmutableStore")
}
