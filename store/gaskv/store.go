package gaskv

import "cosmossdk.io/store/v2"

var _ store.KVStore = (*Store)(nil)

type Store struct {
	parent store.KVStore
}
